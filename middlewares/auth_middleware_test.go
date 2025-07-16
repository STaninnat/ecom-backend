package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
)

type mockLogger struct {
	withErrorCalled bool
	errorCalled     bool
}

func (m *mockLogger) WithError(err error) interface{ Error(args ...any) } {
	m.withErrorCalled = true
	return m
}
func (m *mockLogger) Error(args ...any) {
	m.errorCalled = true
}

// ---
type mockAuthService struct {
	validateFunc func(token, secret string) (*Claims, error)
}

func (m *mockAuthService) ValidateAccessToken(token, secret string) (*Claims, error) {
	return m.validateFunc(token, secret)
}

type mockUserService struct {
	getUserFunc func(ctx context.Context, id string) (database.User, error)
}

func (m *mockUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return m.getUserFunc(ctx, id)
}

type mockMetadataService struct {
	ip string
	ua string
}

func (m *mockMetadataService) GetIPAddress(r *http.Request) string { return m.ip }
func (m *mockMetadataService) GetUserAgent(r *http.Request) string { return m.ua }

func TestLogHandlerError_WithError(t *testing.T) {
	logger := &mockLogger{}
	LogHandlerError(logger, context.Background(), "a", "b", "msg", "ip", "ua", errors.New("fail"))
	if !logger.withErrorCalled || !logger.errorCalled {
		t.Error("expected WithError and Error to be called")
	}
}

func TestLogHandlerError_NoError(t *testing.T) {
	logger := &mockLogger{}
	LogHandlerError(logger, context.Background(), "a", "b", "msg", "ip", "ua", nil)
	if logger.withErrorCalled || !logger.errorCalled {
		t.Error("expected only Error to be called")
	}
}

func TestGetRequestMetadata(t *testing.T) {
	meta := &mockMetadataService{ip: "1.2.3.4", ua: "ua"}
	r := httptest.NewRequest("GET", "/", nil)
	ip, ua := GetRequestMetadata(meta, r)
	if ip != "1.2.3.4" || ua != "ua" {
		t.Errorf("expected 1.2.3.4/ua, got %s/%s", ip, ua)
	}
}

func TestCreateAuthMiddleware_MissingToken(t *testing.T) {
	logger := &mockLogger{}
	mw := CreateAuthMiddleware(&mockAuthService{}, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		t.Error("handler should not be called")
	})
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Code != 401 {
		t.Errorf("expected 401, got %d", rw.Code)
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateAuthMiddleware_InvalidToken(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return nil, errors.New("bad token") }}
	mw := CreateAuthMiddleware(auth, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		t.Error("handler should not be called")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "bad"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Code != 401 {
		t.Errorf("expected 401, got %d", rw.Code)
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateAuthMiddleware_UserLookupFail(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) {
		return database.User{}, errors.New("fail")
	}}
	mw := CreateAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		t.Error("handler should not be called")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Code != 500 {
		t.Errorf("expected 500, got %d", rw.Code)
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateAuthMiddleware_Success(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) { return user, nil }}
	mw := CreateAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		called = true
		if u.ID != "u1" {
			t.Errorf("expected user u1, got %v", u)
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called on success")
	}
}

func TestCreateAdminOnlyMiddleware_NonAdmin(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) { return user, nil }}
	mw := CreateAdminOnlyMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		t.Error("handler should not be called for non-admin")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Code != 403 {
		t.Errorf("expected 403, got %d", rw.Code)
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateAdminOnlyMiddleware_Admin(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "admin"}, nil }}
	user := database.User{ID: "admin", Role: "admin"}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) { return user, nil }}
	mw := CreateAdminOnlyMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u database.User) {
		called = true
		if u.ID != "admin" {
			t.Errorf("expected admin user, got %v", u)
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called for admin")
	}
}

func TestCreateOptionalAuthMiddleware_NoToken(t *testing.T) {
	mw := CreateOptionalAuthMiddleware(&mockAuthService{}, &mockUserService{}, &mockLogger{}, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u *database.User) {
		called = true
		if u != nil {
			t.Error("expected nil user when no token")
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called")
	}
}

func TestCreateOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return nil, errors.New("bad token") }}
	mw := CreateOptionalAuthMiddleware(auth, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u *database.User) {
		called = true
		if u != nil {
			t.Error("expected nil user for invalid token")
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "bad"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called")
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateOptionalAuthMiddleware_UserLookupFail(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) {
		return database.User{}, errors.New("fail")
	}}
	mw := CreateOptionalAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u *database.User) {
		called = true
		if u != nil {
			t.Error("expected nil user for user lookup fail")
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called")
	}
	if !logger.errorCalled {
		t.Error("expected error to be logged")
	}
}

func TestCreateOptionalAuthMiddleware_Success(t *testing.T) {
	auth := &mockAuthService{validateFunc: func(token, secret string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(ctx context.Context, id string) (database.User, error) { return user, nil }}
	mw := CreateOptionalAuthMiddleware(auth, userSvc, &mockLogger{}, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(w http.ResponseWriter, r *http.Request, u *database.User) {
		called = true
		if u == nil || u.ID != "u1" {
			t.Errorf("expected user u1, got %v", u)
		}
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "good"})
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("handler not called on success")
	}
}
