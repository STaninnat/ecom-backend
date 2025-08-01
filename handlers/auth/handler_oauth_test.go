// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
package authhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
)

// handler_oauth_test.go: Test handlers and unit tests for Google OAuth signin and callback flow.

// TestHandlerGoogleSignIn_Success checks that a successful Google sign-in redirects to the correct auth URL.
func TestHandlerGoogleSignIn_Success(t *testing.T) {
	cfg := setupTestConfig()
	expectedAuthURL := "https://accounts.google.com/oauth/authorize?state=test-state&client_id=test-client"

	// Mock successful auth URL generation
	cfg.authService.(*MockAuthService).On("GenerateGoogleAuthURL", mock.Anything).
		Return(expectedAuthURL, nil)

	// Create request
	req := httptest.NewRequest("GET", "/auth/google/signin", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedAuthURL, w.Header().Get("Location"))

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
}

// TestHandlerGoogleSignIn_AuthURLGenerationFailed checks that a failure to generate the Google auth URL returns a 500 error and logs appropriately.
func TestHandlerGoogleSignIn_AuthURLGenerationFailed(t *testing.T) {
	cfg := setupTestConfig()

	// Mock auth URL generation failure
	cfg.authService.(*MockAuthService).On("GenerateGoogleAuthURL", mock.Anything).
		Return("", errors.New("failed to generate URL"))

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-google", "auth_url_generation_failed", "Error generating Google auth URL", mock.Anything, mock.Anything, mock.Anything)

	// Create request
	req := httptest.NewRequest("GET", "/auth/google/signin", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to initiate Google signin")

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_Success checks that a successful Google OAuth callback sets cookies and returns a success response.
func TestHandlerGoogleCallback_Success(t *testing.T) {
	cfg := setupTestConfig()
	userID := "test-user-id"
	accessToken := "test-access-token"
	refreshToken := "test-refresh-token"

	// Mock successful Google auth handling
	cfg.authService.(*MockAuthService).On("HandleGoogleAuth", mock.Anything, "test-code", "test-state").
		Return(&AuthResult{
			UserID:              userID,
			AccessToken:         accessToken,
			RefreshToken:        refreshToken,
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           false,
		}, nil)

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "callback-google", "Google signin success", mock.Anything, mock.Anything)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/auth/google/callback?state=test-state&code=test-code", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Google signin successful")

	// Check that cookies were set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_MissingState checks that a missing state parameter returns a bad request error and logs appropriately.
func TestHandlerGoogleCallback_MissingState(t *testing.T) {
	cfg := setupTestConfig()

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything)

	// Create request with missing state parameter
	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing required parameters")

	// Verify mock expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_MissingCode checks that a missing code parameter returns a bad request error and logs appropriately.
func TestHandlerGoogleCallback_MissingCode(t *testing.T) {
	cfg := setupTestConfig()

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything)

	// Create request with missing code parameter
	req := httptest.NewRequest("GET", "/auth/google/callback?state=test-state", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing required parameters")

	// Verify mock expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_MissingBothParameters checks that missing both state and code returns a bad request error and logs appropriately.
func TestHandlerGoogleCallback_MissingBothParameters(t *testing.T) {
	cfg := setupTestConfig()

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything)

	// Create request with no parameters
	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing required parameters")

	// Verify mock expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_ServiceError checks that a service error during Google callback returns a bad request and logs appropriately.
func TestHandlerGoogleCallback_ServiceError(t *testing.T) {
	cfg := setupTestConfig()

	// Mock service error
	serviceError := &AuthError{Code: "invalid_state", Message: "Invalid state parameter"}
	cfg.authService.(*MockAuthService).On("HandleGoogleAuth", mock.Anything, "test-code", "test-state").
		Return(nil, serviceError)

	// Mock logging (accept any value for the error argument)
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "invalid_state", "Invalid state parameter", mock.Anything, mock.Anything, mock.Anything)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/auth/google/callback?state=test-state&code=test-code", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid state parameter")

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_GenericError checks that a generic error during Google callback returns a 500 error and logs appropriately.
func TestHandlerGoogleCallback_GenericError(t *testing.T) {
	cfg := setupTestConfig()

	// Mock generic error
	genericError := errors.New("some unexpected error")
	cfg.authService.(*MockAuthService).On("HandleGoogleAuth", mock.Anything, "test-code", "test-state").
		Return(nil, genericError)

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, genericError)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/auth/google/callback?state=test-state&code=test-code", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerGoogleCallback(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// Test that the handlers exist and can be called (basic smoke tests)
// TestHandlerGoogleSignIn_Exists is a smoke test to ensure the Google sign-in handler exists and can be called without panicking.
func TestHandlerGoogleSignIn_Exists(t *testing.T) {
	cfg := setupTestConfig()

	// Mock auth URL generation failure (expected for test)
	cfg.authService.(*MockAuthService).On("GenerateGoogleAuthURL", mock.Anything).
		Return("", errors.New("test error"))

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-google", "auth_url_generation_failed", "Error generating Google auth URL", mock.Anything, mock.Anything, mock.Anything)

	// Create request
	req := httptest.NewRequest("GET", "/auth/google/signin", nil)
	w := httptest.NewRecorder()

	// Call the handler - it will fail, but that's expected
	cfg.HandlerGoogleSignIn(w, req)

	// Should return an error status
	assert.NotEqual(t, http.StatusOK, w.Code)

	// Verify mock expectations
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerGoogleCallback_Exists is a smoke test to ensure the Google callback handler exists and can be called without panicking.
func TestHandlerGoogleCallback_Exists(t *testing.T) {
	cfg := setupTestConfig()

	// Mock logging for missing parameters
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything)

	// Create request without parameters
	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	w := httptest.NewRecorder()

	// Call the handler - it will fail due to missing parameters, but that's expected
	cfg.HandlerGoogleCallback(w, req)

	// Should return an error status
	assert.NotEqual(t, http.StatusOK, w.Code)

	// Verify mock expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestRealHandlerGoogleSignIn_AuthURLGenerationFailed checks the real handler for auth URL generation failure and proper error handling.
func TestRealHandlerGoogleSignIn_AuthURLGenerationFailed(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	mockAuthService := &MockAuthService{}
	realAuthConfig := &auth.Config{}

	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			Auth: realAuthConfig,
		},
		Logger:      mockHandlersConfig,
		authService: mockAuthService,
	}

	req := httptest.NewRequest("GET", "/auth/google/signin", nil)
	w := httptest.NewRecorder()

	// Set up mock expectations for the error path
	mockAuthService.On("GenerateGoogleAuthURL", mock.Anything).Return("", errors.New("failed to generate URL"))
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-google", "auth_url_generation_failed", "Error generating Google auth URL", mock.Anything, mock.Anything, mock.Anything).Return()

	cfg.HandlerGoogleSignIn(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to initiate Google signin")
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
}

// TestRealHandler_MissingOrInvalidParameters covers missing/invalid parameter scenarios for real handlers (Google callback and refresh token).
func TestRealHandler_MissingOrInvalidParameters(t *testing.T) {
	tests := []struct {
		name         string
		handlerFunc  func(cfg *HandlersAuthConfig, w http.ResponseWriter, req *http.Request)
		request      *http.Request
		logOp        string
		logCode      string
		logMsg       string
		expectedCode int
		expectedBody string
	}{
		{
			name: "GoogleCallback_MissingParameters",
			handlerFunc: func(cfg *HandlersAuthConfig, w http.ResponseWriter, req *http.Request) {
				cfg.HandlerGoogleCallback(w, req)
			},
			request:      httptest.NewRequest("GET", "/auth/google/callback", nil),
			logOp:        "callback-google",
			logCode:      "missing_parameters",
			logMsg:       "Missing state or code parameter",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Missing required parameters",
		},
		{
			name: "RefreshToken_InvalidToken",
			handlerFunc: func(cfg *HandlersAuthConfig, w http.ResponseWriter, req *http.Request) {
				cfg.HandlerRefreshToken(w, req)
			},
			request:      httptest.NewRequest("POST", "/refresh", nil),
			logOp:        "refresh_token",
			logCode:      "invalid_token",
			logMsg:       "Error validating authentication token",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "http: named cookie not present",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockHandlersConfig := &MockHandlersConfig{}
			mockAuthService := &MockAuthService{}
			realAuthConfig := &auth.Config{}

			cfg := &HandlersAuthConfig{
				Config: &handlers.Config{
					Auth: realAuthConfig,
				},
				Logger:      mockHandlersConfig,
				authService: mockAuthService,
			}

			mockHandlersConfig.On("LogHandlerError", mock.Anything, tc.logOp, tc.logCode, tc.logMsg, mock.Anything, mock.Anything, mock.Anything).Return()

			w := httptest.NewRecorder()
			tc.handlerFunc(cfg, w, tc.request)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
			mockHandlersConfig.AssertExpectations(t)
		})
	}
}
