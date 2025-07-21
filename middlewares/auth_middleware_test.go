// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// auth_middleware_test.go: Tests for authentication and authorization middleware.

type mockLogger struct {
	withErrorCalled bool
	errorCalled     bool
}

// WithError mocks the logger WithError method for testing purposes
func (m *mockLogger) WithError(_ error) interface{ Error(args ...any) } {
	m.withErrorCalled = true
	return m
}
func (m *mockLogger) Error(_ ...any) {
	m.errorCalled = true
}

// ---
type mockAuthService struct {
	validateFunc func(token, secret string) (*Claims, error)
}

// ValidateAccessToken mocks the auth service token validation for testing purposes
func (m *mockAuthService) ValidateAccessToken(token, secret string) (*Claims, error) {
	return m.validateFunc(token, secret)
}

type mockUserService struct {
	getUserFunc func(ctx context.Context, id string) (database.User, error)
}

// GetUserByID mocks the user service user lookup for testing purposes
func (m *mockUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return m.getUserFunc(ctx, id)
}

type mockMetadataService struct {
	ip string
	ua string
}

// GetIPAddress mocks the metadata service IP extraction for testing purposes
func (m *mockMetadataService) GetIPAddress(_ *http.Request) string { return m.ip }

// GetUserAgent mocks the metadata service user agent extraction for testing purposes
func (m *mockMetadataService) GetUserAgent(_ *http.Request) string { return m.ua }

// TestLogHandlerError_WithError tests error logging when an error is provided
// It verifies that both WithError and Error methods are called on the logger
func TestLogHandlerError_WithError(t *testing.T) {
	logger := &mockLogger{}
	LogHandlerError(context.Background(), logger, "a", "b", "msg", "ip", "ua", errors.New("fail"))
	if !logger.withErrorCalled || !logger.errorCalled {
		t.Error("expected WithError and Error to be called")
	}
}

// TestLogHandlerError_NoError tests error logging when no error is provided
// It verifies that only the Error method is called, not WithError
func TestLogHandlerError_NoError(t *testing.T) {
	logger := &mockLogger{}
	LogHandlerError(context.Background(), logger, "a", "b", "msg", "ip", "ua", nil)
	if logger.withErrorCalled || !logger.errorCalled {
		t.Error("expected only Error to be called")
	}
}

// TestGetRequestMetadata tests metadata extraction from requests
// It verifies that IP address and user agent are correctly extracted from the metadata service
func TestGetRequestMetadata(t *testing.T) {
	meta := &mockMetadataService{ip: "1.2.3.4", ua: "ua"}
	r := httptest.NewRequest("GET", "/", nil)
	ip, ua := GetRequestMetadata(meta, r)
	if ip != "1.2.3.4" || ua != "ua" {
		t.Errorf("expected 1.2.3.4/ua, got %s/%s", ip, ua)
	}
}

// TestCreateAuthMiddleware_MissingToken tests authentication when no token is provided
// It verifies that the middleware returns 401 and logs the missing token error
func TestCreateAuthMiddleware_MissingToken(t *testing.T) {
	logger := &mockLogger{}
	mw := CreateAuthMiddleware(&mockAuthService{}, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	h := mw(func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
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

// TestCreateAuthMiddleware_InvalidToken tests authentication with an invalid token
// It verifies that the middleware returns 401 and logs the token validation error
func TestCreateAuthMiddleware_InvalidToken(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return nil, errors.New("bad token") }}
	mw := CreateAuthMiddleware(auth, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	h := mw(func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
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

// TestCreateAuthMiddleware_UserLookupFail tests authentication when user lookup fails
// It verifies that the middleware returns 500 and logs the user lookup error
func TestCreateAuthMiddleware_UserLookupFail(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) {
		return database.User{}, errors.New("fail")
	}}
	mw := CreateAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	h := mw(func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
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

// TestCreateAuthMiddleware_Success tests successful authentication flow
// It verifies that the middleware calls the handler with the correct user when authentication succeeds
func TestCreateAuthMiddleware_Success(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) { return user, nil }}
	mw := CreateAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u database.User) {
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

// TestCreateAdminOnlyMiddleware_NonAdmin tests admin middleware with non-admin users
// It verifies that non-admin users are denied access with 403 status and error logging
func TestCreateAdminOnlyMiddleware_NonAdmin(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) { return user, nil }}
	mw := CreateAdminOnlyMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	h := mw(func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
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

// TestCreateAdminOnlyMiddleware_Admin tests admin middleware with admin users
// It verifies that admin users are allowed access and the handler is called with the admin user
func TestCreateAdminOnlyMiddleware_Admin(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "admin"}, nil }}
	user := database.User{ID: "admin", Role: "admin"}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) { return user, nil }}
	mw := CreateAdminOnlyMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u database.User) {
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

// TestCreateOptionalAuthMiddleware_NoToken tests optional auth when no token is provided
// It verifies that the handler is called with nil user when no authentication token exists
func TestCreateOptionalAuthMiddleware_NoToken(t *testing.T) {
	mw := CreateOptionalAuthMiddleware(&mockAuthService{}, &mockUserService{}, &mockLogger{}, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u *database.User) {
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

// TestCreateOptionalAuthMiddleware_InvalidToken tests optional auth with invalid token
// It verifies that the handler is called with nil user and errors are logged for invalid tokens
func TestCreateOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return nil, errors.New("bad token") }}
	mw := CreateOptionalAuthMiddleware(auth, &mockUserService{}, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u *database.User) {
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

// TestCreateOptionalAuthMiddleware_UserLookupFail tests optional auth when user lookup fails
// It verifies that the handler is called with nil user and errors are logged for lookup failures
func TestCreateOptionalAuthMiddleware_UserLookupFail(t *testing.T) {
	logger := &mockLogger{}
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) {
		return database.User{}, errors.New("fail")
	}}
	mw := CreateOptionalAuthMiddleware(auth, userSvc, logger, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u *database.User) {
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

// TestCreateOptionalAuthMiddleware_Success tests successful optional authentication flow
// It verifies that the middleware calls the handler with the correct user when authentication succeeds
func TestCreateOptionalAuthMiddleware_Success(t *testing.T) {
	auth := &mockAuthService{validateFunc: func(_, _ string) (*Claims, error) { return &Claims{UserID: "u1"}, nil }}
	user := database.User{ID: "u1", Role: "user"}
	userSvc := &mockUserService{getUserFunc: func(_ context.Context, _ string) (database.User, error) { return user, nil }}
	mw := CreateOptionalAuthMiddleware(auth, userSvc, &mockLogger{}, &mockMetadataService{}, "secret")
	called := false
	h := mw(func(_ http.ResponseWriter, _ *http.Request, u *database.User) {
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
