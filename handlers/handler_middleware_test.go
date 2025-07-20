package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of the AuthService interface for testing.
type MockAuthService struct {
	mock.Mock
}

// ValidateAccessToken mocks the ValidateAccessToken method for AuthService.
func (m *MockAuthService) ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	args := m.Called(tokenString, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

// GenerateAccessToken mocks the GenerateAccessToken method for AuthService.
func (m *MockAuthService) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	args := m.Called(userID, expiresAt)
	return args.String(0), args.Error(1)
}

// GenerateRefreshToken mocks the GenerateRefreshToken method for AuthService.
func (m *MockAuthService) GenerateRefreshToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// GenerateTokens mocks the GenerateTokens method for AuthService.
func (m *MockAuthService) GenerateTokens(userID string, accessTokenExpiresAt time.Time) (string, string, error) {
	args := m.Called(userID, accessTokenExpiresAt)
	return args.String(0), args.String(1), args.Error(2)
}

// StoreRefreshTokenInRedis mocks the StoreRefreshTokenInRedis method for AuthService.
func (m *MockAuthService) StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error {
	args := m.Called(r, userID, refreshToken, provider, ttl)
	return args.Error(0)
}

// ValidateRefreshToken mocks the ValidateRefreshToken method for AuthService.
func (m *MockAuthService) ValidateRefreshToken(refreshToken string) (string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.Error(1)
}

// ValidateCookieRefreshTokenData mocks the ValidateCookieRefreshTokenData method for AuthService.
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

// MockUserService is a mock implementation of the UserService interface for testing.
type MockUserService struct {
	mock.Mock
}

// GetUserByID mocks the GetUserByID method for UserService.
func (m *MockUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.User), args.Error(1)
}

// GetUserByEmail mocks the GetUserByEmail method for UserService.
func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

// CheckUserExistsByEmail mocks the CheckUserExistsByEmail method for UserService.
func (m *MockUserService) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// CheckUserExistsByName mocks the CheckUserExistsByName method for UserService.
func (m *MockUserService) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// CreateUser mocks the CreateUser method for UserService.
func (m *MockUserService) CreateUser(ctx context.Context, arg database.CreateUserParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// UpdateUserInfo mocks the UpdateUserInfo method for UserService.
func (m *MockUserService) UpdateUserInfo(ctx context.Context, arg database.UpdateUserInfoParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// UpdateUserRole mocks the UpdateUserRole method for UserService.
func (m *MockUserService) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// CheckExistsAndGetIDByEmail mocks the CheckExistsAndGetIDByEmail method for UserService.
func (m *MockUserService) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.CheckExistsAndGetIDByEmailRow), args.Error(1)
}

// TestHandlerConfig_HandlerMiddleware_Success tests the HandlerMiddleware for successful authentication and user retrieval.
// It checks that the handler is called and returns status OK when all dependencies succeed.
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

// TestHandlerConfig_HandlerMiddleware_MissingToken tests HandlerMiddleware when the access token cookie is missing.
// It checks that the handler is not called and returns status Unauthorized.
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

// TestHandlerConfig_HandlerMiddleware_InvalidToken tests HandlerMiddleware with an invalid access token.
// It checks that the handler is not called and returns status Unauthorized.
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

// TestHandlerConfig_HandlerMiddleware_UserNotFound tests HandlerMiddleware when the user is not found.
// It checks that the handler is not called and returns status InternalServerError.
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

// TestHandlerConfig_HandlerAdminOnlyMiddleware_AdminUser tests HandlerAdminOnlyMiddleware for an admin user.
// It checks that the handler is called and returns status OK for admin users.
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

// TestHandlerConfig_HandlerAdminOnlyMiddleware_NonAdminUser tests HandlerAdminOnlyMiddleware for a non-admin user.
// It checks that the handler is not called and returns status Forbidden.
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

// TestHandlerAdminOnlyMiddleware_UnknownRole tests HandlerAdminOnlyMiddleware for a user with an unknown role.
// It checks that the handler is not called and returns status Forbidden.
func TestHandlerAdminOnlyMiddleware_UnknownRole(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "admin-token"})
	w := httptest.NewRecorder()

	expectedUser := database.User{ID: "user999", Role: "unknown"}
	expectedClaims := &Claims{UserID: "user999"}

	mockRequestMetadata.On("GetIPAddress", req).Return("10.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")
	mockAuth.On("ValidateAccessToken", "admin-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user999").Return(expectedUser, nil)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(nil)
	mockLogger.On("WithError", mock.Anything).Return(entry).Maybe()
	mockLogger.On("Error", mock.Anything).Return().Maybe()

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
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerMiddleware_HandlerWritesError tests HandlerMiddleware when the handler writes an error response.
// It checks that the response status matches the error code.
func TestHandlerMiddleware_HandlerWritesError(t *testing.T) {
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

	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		http.Error(w, "handler error", http.StatusTeapot)
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTeapot, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerMiddleware_TokenExpired tests HandlerMiddleware with an expired token.
// It checks that the handler is not called and returns status Unauthorized.
func TestHandlerMiddleware_TokenExpired(t *testing.T) {
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
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "expired-token"})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "expired-token", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	testHandler := AuthHandler(func(w http.ResponseWriter, r *http.Request, user database.User) {
		// Should not be called
		assert.Fail(t, "handler should not be called for expired token")
	})

	middleware := cfg.HandlerMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerMiddleware_LoggerServiceNil tests HandlerMiddleware with a nil LoggerService.
// It checks that the handler is called and does not panic.
func TestHandlerMiddleware_LoggerServiceNil(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
		LoggerService:          nil, // Intentionally nil
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
	assert.NotPanics(t, func() {
		middleware.ServeHTTP(w, req)
	})
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// Add custom type for context key
type testKeyType string

const testKey testKeyType = "test-key"

// TestHandlerOptionalMiddleware_Success tests HandlerOptionalMiddleware for successful authentication.
// It checks that the handler is called with a non-nil user and returns status OK.
func TestHandlerOptionalMiddleware_Success(t *testing.T) {
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
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, *user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerOptionalMiddleware_MissingToken tests HandlerOptionalMiddleware when the access token is missing.
// It checks that the handler is called with a nil user and returns status OK.
func TestHandlerOptionalMiddleware_MissingToken(t *testing.T) {
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

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerOptionalMiddleware_InvalidToken tests HandlerOptionalMiddleware with an invalid token.
// It checks that the handler is called with a nil user and returns status OK.
func TestHandlerOptionalMiddleware_InvalidToken(t *testing.T) {
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

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerOptionalMiddleware_UserNotFound tests HandlerOptionalMiddleware when the user is not found.
// It checks that the handler is called with a nil user and returns status OK.
func TestHandlerOptionalMiddleware_UserNotFound(t *testing.T) {
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

	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(database.User{}, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlersConfig_HandlerOptionalMiddleware_Structure tests the structure of HandlersConfig for optional middleware.
// It checks that the adapter can be created (structure test).
func TestHandlersConfig_HandlerOptionalMiddleware_Structure(t *testing.T) {
	adapter := &HandlersConfig{}
	assert.NotNil(t, adapter)
}

// TestHandlerOptionalMiddleware_MalformedCookie tests HandlerOptionalMiddleware with a malformed or empty cookie.
// It checks that the handler is called with a nil user and returns status OK.
func TestHandlerOptionalMiddleware_MalformedCookie(t *testing.T) {
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
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""}) // Malformed/empty
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")
	mockAuth.On("ValidateAccessToken", "", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerOptionalMiddleware_MultipleCookies tests HandlerOptionalMiddleware with multiple cookies.
// It checks that the handler is called with a nil user and returns status OK.
func TestHandlerOptionalMiddleware_MultipleCookies(t *testing.T) {
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
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")
	mockAuth.On("ValidateAccessToken", "", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlerOptionalMiddleware_ContextPropagation tests context propagation in HandlerOptionalMiddleware.
// It checks that the context value is preserved in the handler.
func TestHandlerOptionalMiddleware_ContextPropagation(t *testing.T) {
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}
	cfg := &HandlerConfig{
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), testKey, "test-value")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")

	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		assert.Equal(t, "test-value", r.Context().Value(testKey))
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)
	mockRequestMetadata.AssertExpectations(t)
}

// TestHandlersConfig_HandlerAdminOnlyMiddleware tests the legacy HandlerAdminOnlyMiddleware.
// It checks that the middleware does not panic and returns a valid response code.
func TestHandlersConfig_HandlerAdminOnlyMiddleware(t *testing.T) {
	logger := logrus.New()
	apiCfg := &config.APIConfig{JWTSecret: "test-secret", DB: &database.Queries{}}
	cfg := &HandlersConfig{
		APIConfig: apiCfg,
		Auth:      &auth.AuthConfig{},
		Logger:    logger,
	}

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "admin-token"})
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request, user database.User) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := cfg.HandlerAdminOnlyMiddleware(handler)
	assert.NotNil(t, middleware)
	middleware(w, req)
	// We can't assert handler invocation without a real DB and token, but we can check for no panic and a valid response
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized)
}

// TestHandlersConfig_HandlerMiddleware tests the legacy HandlerMiddleware.
// It checks that the middleware does not panic and returns a valid response code.
func TestHandlersConfig_HandlerMiddleware(t *testing.T) {
	logger := logrus.New()
	apiCfg := &config.APIConfig{JWTSecret: "test-secret", DB: &database.Queries{}}
	cfg := &HandlersConfig{
		APIConfig: apiCfg,
		Auth:      &auth.AuthConfig{},
		Logger:    logger,
	}

	req := httptest.NewRequest("GET", "/user", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "user-token"})
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request, user database.User) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := cfg.HandlerMiddleware(handler)
	assert.NotNil(t, middleware)
	middleware(w, req)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized)
}

// TestHandlersConfig_HandlerOptionalMiddleware tests the legacy HandlerOptionalMiddleware.
// It checks that the middleware does not panic and returns a valid response code.
func TestHandlersConfig_HandlerOptionalMiddleware(t *testing.T) {
	logger := logrus.New()
	apiCfg := &config.APIConfig{JWTSecret: "test-secret", DB: &database.Queries{}}
	cfg := &HandlersConfig{
		APIConfig: apiCfg,
		Auth:      &auth.AuthConfig{},
		Logger:    logger,
	}

	req := httptest.NewRequest("GET", "/optional", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "user-token"})
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request, user *database.User) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := cfg.HandlerOptionalMiddleware(handler)
	assert.NotNil(t, middleware)
	middleware(w, req)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized)
}
