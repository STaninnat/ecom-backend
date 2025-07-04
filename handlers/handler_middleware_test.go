package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	args := m.Called(tokenString, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockAuthService) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	args := m.Called(userID, expiresAt)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateRefreshToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateTokens(userID string, accessTokenExpiresAt time.Time) (string, string, error) {
	args := m.Called(userID, accessTokenExpiresAt)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error {
	args := m.Called(r, userID, refreshToken, provider, ttl)
	return args.Error(0)
}

func (m *MockAuthService) ValidateRefreshToken(refreshToken string) (string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (string, *RefreshTokenData, error) {
	args := m.Called(w, r)
	if args.Get(0) == nil {
		return "", nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*RefreshTokenData), args.Error(2)
}

// MockUserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserService) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, arg database.CreateUserParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockUserService) UpdateUserInfo(ctx context.Context, arg database.UpdateUserInfoParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockUserService) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockUserService) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.CheckExistsAndGetIDByEmailRow), args.Error(1)
}

func TestHandlerConfig_HandlerMiddleware_Success(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "test-token"})
	w := httptest.NewRecorder()

	expectedUser := database.User{ID: "user123", Role: "user"}
	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(expectedUser, nil)

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
		assert.Equal(t, expectedUser, user)
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_HandlerMiddleware_MissingToken(t *testing.T) {
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockLogger.On("WithError", mock.MatchedBy(func(err error) bool {
		return err != nil && err.Error() == "http: named cookie not present"
	})).Return(logrus.New().WithField("test", "test"))
	mockLogger.On("Error", "Access token cookie not found").Return()

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_HandlerMiddleware_InvalidToken(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "invalid-token", "test-secret").Return(nil, assert.AnError)
	mockLogger.On("WithError", mock.MatchedBy(func(err error) bool {
		return err != nil && err.Error() == assert.AnError.Error()
	})).Return(logrus.New().WithField("test", "test"))
	mockLogger.On("Error", "Access token validation failed").Return()

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuth.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_HandlerMiddleware_UserNotFound(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	token := "valid-token"
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", token, "test-secret").Return(&Claims{UserID: "user-id"}, nil)
	mockUser.On("GetUserByID", mock.Anything, "user-id").Return(database.User{}, assert.AnError)
	mockLogger.On("WithError", mock.MatchedBy(func(err error) bool {
		return err != nil && err.Error() == assert.AnError.Error()
	})).Return(logrus.New().WithField("test", "test"))
	mockLogger.On("Error", "User not found").Return()

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_HandlerAdminOnlyMiddleware_AdminUser(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "test-token"})
	w := httptest.NewRecorder()

	expectedUser := database.User{ID: "user123", Role: "admin"}
	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(expectedUser, nil)

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
		assert.Equal(t, expectedUser, user)
	})

	middleware := cfg.HandlerAdminOnlyMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_HandlerAdminOnlyMiddleware_NonAdminUser(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "test-token"})
	w := httptest.NewRecorder()

	expectedUser := database.User{ID: "user123", Role: "user"}
	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(expectedUser, nil)
	mockLogger.On("Error", "unauthorized access attempt").Return()

	handlerCalled := false
	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		handlerCalled = true
	})

	middleware := cfg.HandlerAdminOnlyMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}
