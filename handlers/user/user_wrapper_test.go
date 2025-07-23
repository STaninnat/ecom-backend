// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	authpkg "github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// user_wrapper_test.go: Tests for thread-safe user service initialization, error handling, user extraction middleware, and auth handlers.

// TestInitUserService_MissingHandlersConfig tests that InitUserService returns an error
// when the Config is nil
func TestInitUserService_MissingHandlersConfig(t *testing.T) {
	cfg := &HandlersUserConfig{Config: nil}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

// TestInitUserService_MissingDB tests that InitUserService returns an error
// when the database is nil
func TestInitUserService_MissingDB(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{APIConfig: &config.APIConfig{DB: nil}}}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestGetUserService_AlreadyInitialized tests that GetUserService returns
// the existing userService when it's already initialized
func TestGetUserService_AlreadyInitialized(t *testing.T) {
	mockService := new(MockUserService)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{},
		userService: mockService,
	}
	service := cfg.GetUserService()
	assert.Equal(t, mockService, service)
}

// TestGetUserService_InitializesWithNilConfig tests that GetUserService
// initializes a new service even when Config is nil
func TestGetUserService_InitializesWithNilConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config:      nil,
		userService: nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

// TestGetUserService_InitializesWithNilDB tests that GetUserService
// initializes a new service even when database is nil
func TestGetUserService_InitializesWithNilDB(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{APIConfig: &config.APIConfig{DB: nil}},
		userService: nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

// TestGetUserService_InitializesWithValidConfig tests that GetUserService
// initializes a new service with valid configuration
func TestGetUserService_InitializesWithValidConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{DB: &database.Queries{}},
		},
		userService: nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

// --- HandleUserError ---

// TestHandleUserError_KnownError tests that handleUserError correctly handles
// known error types like update_failed
func TestHandleUserError_KnownError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "update_failed", Message: "fail", Err: errors.New("db")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "update_failed", "fail", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_UnknownError tests that handleUserError correctly handles
// unknown error types
func TestHandleUserError_UnknownError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := errors.New("unknown error")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_CommitError tests that handleUserError correctly handles
// commit_error type errors
func TestHandleUserError_CommitError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "commit_error", Message: "Commit failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "commit_error", "Commit failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_TransactionError tests that handleUserError correctly handles
// transaction_error type errors
func TestHandleUserError_TransactionError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "transaction_error", Message: "Transaction failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "transaction_error", "Transaction failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

// TestInitUserService_Success tests that InitUserService successfully initializes
// the userService with valid configuration
func TestInitUserService_Success(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{DB: &database.Queries{}},
		},
		Logger: new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.NoError(t, err)
	assert.NotNil(t, cfg.userService)
}

// TestHandleUserError_DefaultCase tests that handleUserError correctly handles
// unknown AppError codes by defaulting to internal_error
func TestHandleUserError_DefaultCase(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "other_error", Message: "fail", Err: errors.New("db")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "internal_error", "fail", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_NonUserError tests that handleUserError correctly handles
// non-AppError types by treating them as unknown errors
func TestHandleUserError_NonUserError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := errors.New("unexpected error")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_UserNotFound tests that handleUserError correctly handles
// user_not_found errors with 404 status
func TestHandleUserError_UserNotFound(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "user_not_found", Message: "User not found", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "user_not_found", "User not found", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusNotFound, w.status)
	assert.Contains(t, w.body, "User not found")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_InvalidRequest tests that handleUserError correctly handles
// invalid_request errors with 400 status
func TestHandleUserError_InvalidRequest(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid request", Err: errors.New("validation error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "invalid_request", "Invalid request", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusBadRequest, w.status)
	assert.Contains(t, w.body, "Invalid request")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_UpdateFailed tests that handleUserError correctly handles
// update_failed errors
func TestHandleUserError_UpdateFailed(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "update_failed", Message: "Update failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "update_failed", "Update failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

// TestHandleUserError_DefaultAppError tests that handleUserError correctly handles
// custom AppError codes by defaulting to internal_error
func TestHandleUserError_DefaultAppError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "custom_error", Message: "Custom error", Err: errors.New("custom error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "internal_error", "Custom error", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// TestInitUserService_NilHandlersConfig tests that InitUserService returns an error
// when Config is nil
func TestInitUserService_NilHandlersConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config: nil,
		Logger: new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

// TestInitUserService_NilDB tests that InitUserService returns an error
// when database is nil
func TestInitUserService_NilDB(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{APIConfig: &config.APIConfig{DB: nil}},
		Logger: new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestExtractUserFromRequest_NoAuthHeader tests that extractUserFromRequest returns
// an error when no Authorization header is present
func TestExtractUserFromRequest_NoAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request without Authorization header
	r := httptest.NewRequest("GET", "/user", nil)

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

// TestExtractUserFromRequest_EmptyAuthHeader tests that extractUserFromRequest returns
// an error when Authorization header is empty
func TestExtractUserFromRequest_EmptyAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with empty Authorization header
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

// TestExtractUserFromRequest_ShortAuthHeader tests that extractUserFromRequest returns
// an error when Authorization header is too short
func TestExtractUserFromRequest_ShortAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with short Authorization header
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Short")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

// TestExtractUserFromRequest_InvalidBearerPrefix tests that extractUserFromRequest returns
// an error when Authorization header doesn't start with "Bearer "
func TestExtractUserFromRequest_InvalidBearerPrefix(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with invalid Bearer prefix
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "InvalidPrefix valid-token")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

// TestUserExtractionMiddleware_Behavior tests that UserExtractionMiddleware returns
// a valid handler function and is callable
func TestUserExtractionMiddleware_Behavior(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger:      mockLogger,
		userService: mockUserService,
	}

	// Test that the middleware returns a handler function
	middleware := cfg.UserExtractionMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	assert.NotNil(t, middleware)

	// Test that it's callable
	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, r)
	// Should not panic and should complete successfully
}

// TestExtractUserFromRequest_SimpleScenarios tests extractUserFromRequest with
// various simple scenarios that don't require complex mocking
func TestExtractUserFromRequest_SimpleScenarios(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Test cases that don't require auth mocking
	testCases := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "No Authorization header",
			authHeader:    "",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Empty Authorization header",
			authHeader:    "",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Short Authorization header",
			authHeader:    "Short",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Invalid Bearer prefix",
			authHeader:    "InvalidPrefix valid-token",
			expectedError: "missing or invalid Authorization header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/user", nil)
			if tc.authHeader != "" {
				r.Header.Set("Authorization", tc.authHeader)
			}

			user, err := cfg.extractUserFromRequest(r)
			assert.Error(t, err)
			assert.Equal(t, database.User{}, user)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestAuthHandlerGetUser_CallsUnderlyingHandlerWithUserInContext tests that
// AuthHandlerGetUser calls the underlying handler with user in context
func TestAuthHandlerGetUser_CallsUnderlyingHandlerWithUserInContext(t *testing.T) {
	cfg := &testUserConfig{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	user := database.User{ID: "u1", Name: "Test User"}
	cfg.AuthHandlerGetUser(w, r, user)
	assert.True(t, cfg.calledGetUser, "HandlerGetUser should be called")
	assert.Equal(t, user, cfg.gotUser, "User should be injected into context")
}

// TestAuthHandlerUpdateUser_CallsUnderlyingHandlerWithUserInContext tests that
// AuthHandlerUpdateUser calls the underlying handler with user in context
func TestAuthHandlerUpdateUser_CallsUnderlyingHandlerWithUserInContext(t *testing.T) {
	cfg := &testUserConfig{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/users", nil)
	user := database.User{ID: "u2", Name: "Update User"}
	cfg.AuthHandlerUpdateUser(w, r, user)
	assert.True(t, cfg.calledUpdateUser, "HandlerUpdateUser should be called")
	assert.Equal(t, user, cfg.gotUser, "User should be injected into context")
}

// TestAuthHandlerPromoteUserToAdmin_CallsUnderlyingHandlerWithUserInContext tests that
// AuthHandlerPromoteUserToAdmin calls the underlying handler with user in context
func TestAuthHandlerPromoteUserToAdmin_CallsUnderlyingHandlerWithUserInContext(t *testing.T) {
	cfg := &testUserConfig{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/admin/user/promote", nil)
	user := database.User{ID: "u3", Name: "Admin User"}
	cfg.AuthHandlerPromoteUserToAdmin(w, r, user)
	assert.True(t, cfg.calledPromoteAdmin, "HandlerPromoteUserToAdmin should be called")
	assert.Equal(t, user, cfg.gotUser, "User should be injected into context")
}

// ... implement other methods as no-ops if needed

// TestUserExtractionMiddleware_HappyPath tests the happy path of UserExtractionMiddleware
// using real AuthConfig and valid JWT tokens
func TestUserExtractionMiddleware_HappyPath(t *testing.T) {
	// Setup real AuthConfig with a valid JWT
	apiCfg := &config.APIConfig{JWTSecret: "testsecret", Issuer: "test-issuer", Audience: "test-audience"}
	authCfg := &authpkg.Config{APIConfig: apiCfg}
	mockDB := new(MockUserServiceForMiddleware)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      authCfg,
		},
		userService: mockDB,
	}

	// Create a valid JWT
	claims := &authpkg.Claims{
		UserID: "u1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Audience:  []string{"test-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("testsecret"))
	assert.NoError(t, err)

	mockDB.On("GetUserByID", mock.Anything, "u1").Return(database.User{ID: "u1"}, nil)

	called := false
	h := cfg.UserExtractionMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(contextKeyUser).(database.User)
		assert.True(t, ok)
		assert.Equal(t, "u1", user.ID)
		called = true
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)
	h.ServeHTTP(w, r)
	assert.True(t, called)
	mockDB.AssertExpectations(t)
}

// TestExtractUserFromRequest_InvalidToken tests that extractUserFromRequest returns
// an error when the JWT token is invalid (wrong secret)
func TestExtractUserFromRequest_InvalidToken(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)
	apiCfg := &config.APIConfig{JWTSecret: "secret", Issuer: "issuer", Audience: "aud"}
	authCfg := &authpkg.Config{APIConfig: apiCfg}
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      authCfg,
		},
		userService: mockUserService,
	}

	// Generate a JWT with the wrong secret
	claims := &authpkg.Claims{
		UserID: "u1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Audience:  []string{"aud"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("wrongsecret"))
	assert.NoError(t, err)

	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "invalid token")
}

// TestExtractUserFromRequest_ExpiredToken tests that extractUserFromRequest returns
// an error when the JWT token is expired
func TestExtractUserFromRequest_ExpiredToken(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)
	apiCfg := &config.APIConfig{JWTSecret: "secret", Issuer: "issuer", Audience: "aud"}
	authCfg := &authpkg.Config{APIConfig: apiCfg}
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      authCfg,
		},
		userService: mockUserService,
	}

	// Generate an expired JWT
	claims := &authpkg.Claims{
		UserID: "u1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Audience:  []string{"aud"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "invalid token")
}

// TestExtractUserFromRequest_UserNotFound tests that extractUserFromRequest returns
// an error when the user is not found in the database
func TestExtractUserFromRequest_UserNotFound(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)
	apiCfg := &config.APIConfig{JWTSecret: "secret", Issuer: "issuer", Audience: "aud"}
	authCfg := &authpkg.Config{APIConfig: apiCfg}
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      authCfg,
		},
		userService: mockUserService,
	}

	// Generate a valid JWT
	claims := &authpkg.Claims{
		UserID: "u1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Audience:  []string{"aud"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)

	mockUserService.On("GetUserByID", mock.Anything, "u1").Return(database.User{}, errors.New("user not found")).Once()

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "user not found")
	mockUserService.AssertExpectations(t)
}

// TestGetUserService_ConcurrentAccess tests that GetUserService is thread-safe
// and always returns the same instance under concurrent access
func TestGetUserService_ConcurrentAccess(t *testing.T) {
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{APIConfig: &config.APIConfig{}},
		userService: nil,
	}
	var wg sync.WaitGroup
	serviceSet := make(map[UserService]struct{})
	mu := sync.Mutex{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc := cfg.GetUserService()
			mu.Lock()
			serviceSet[svc] = struct{}{}
			mu.Unlock()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, len(serviceSet), "GetUserService should always return the same instance")
}

// TestHandleUserError_AllAppErrorCodes tests that handleUserError correctly handles
// all known AppError codes with appropriate status codes and responses
func TestHandleUserError_AllAppErrorCodes(t *testing.T) {
	codes := []struct {
		code     string
		message  string
		expected int
		response string
	}{
		{"transaction_error", "Transaction failed", http.StatusInternalServerError, "Something went wrong"},
		{"update_failed", "Update failed", http.StatusInternalServerError, "Something went wrong"},
		{"commit_error", "Commit failed", http.StatusInternalServerError, "Something went wrong"},
		{"user_not_found", "User not found", http.StatusNotFound, "User not found"},
		{"invalid_request", "Invalid request", http.StatusBadRequest, "Invalid request"},
		{"other_error", "Other error", http.StatusInternalServerError, "Internal server error"},
	}
	for _, tc := range codes {
		t.Run(tc.code, func(t *testing.T) {
			mockLogger := new(mockHandlerLogger)
			cfg := &HandlersUserConfig{
				Config: &handlers.Config{
					Logger:    logrus.New(),
					APIConfig: &config.APIConfig{},
				},
				Logger: mockLogger,
			}
			err := &handlers.AppError{Code: tc.code, Message: tc.message, Err: errors.New("err")}
			w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
			r := httptest.NewRequest("POST", "/", nil)
			mockLogger.On("LogHandlerError", mock.Anything, "op", mock.Anything, tc.message, "ip", "ua", err.Err).Return()
			cfg.handleUserError(w, r, err, "op", "ip", "ua")
			assert.Equal(t, tc.expected, w.status)
			assert.Contains(t, w.body, tc.response)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandleUserError_UnknownErrorType tests that handleUserError correctly handles
// unknown error types by treating them as unknown errors
func TestHandleUserError_UnknownErrorType(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		Config: &handlers.Config{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := errors.New("some unknown error")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()
	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// TestUserExtractionMiddleware_ErrorPath tests that UserExtractionMiddleware handles
// errors correctly and doesn't set user in context on error
func TestUserExtractionMiddleware_ErrorPath(t *testing.T) {
	cfg := &testUserExtractionConfig{}
	called := false
	h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(contextKeyUser).(database.User)
		assert.False(t, ok, "User should not be set in context on error")
		called = true
	})
	mw := cfg.UserExtractionMiddleware(h)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	mw.ServeHTTP(w, r)
	assert.True(t, called)
}

// TestInitUserService_DirectCall tests InitUserService with a direct call
// to ensure the function is covered
func TestInitUserService_DirectCall(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{APIConfig: &config.APIConfig{DB: &database.Queries{}}}}
	err := cfg.InitUserService()
	assert.NoError(t, err)
}

// TestGetUserService_DirectCall tests GetUserService with a direct call
// to ensure the function is covered
func TestGetUserService_DirectCall(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{APIConfig: &config.APIConfig{DB: &database.Queries{}}}}
	svc := cfg.GetUserService()
	assert.NotNil(t, svc)
}

// TestHandleUserError_DirectCall tests handleUserError with a direct call
// to ensure the function is covered
func TestHandleUserError_DirectCall(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{Logger: logrus.New(), APIConfig: &config.APIConfig{}}, Logger: &dummyHandlerLogger{}}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := errors.New("test error")
	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.NotEqual(t, 0, w.Code)
}

// TestUserExtractionMiddleware_DirectCall tests UserExtractionMiddleware with a direct call
// to ensure the function is covered
func TestUserExtractionMiddleware_DirectCall(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{APIConfig: &config.APIConfig{}}}
	h := cfg.UserExtractionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestExtractUserFromRequest_DirectCall tests extractUserFromRequest with a direct call
// to ensure the function is covered
func TestExtractUserFromRequest_DirectCall(t *testing.T) {
	cfg := &HandlersUserConfig{Config: &handlers.Config{APIConfig: &config.APIConfig{}}}
	r := httptest.NewRequest("GET", "/", nil)
	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
}
