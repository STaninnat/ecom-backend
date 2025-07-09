package authhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// HandlerGoogleSignIn is a test implementation that uses the mocked dependencies
func (cfg *TestHandlersAuthConfig) HandlerGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	// Generate state and auth URL
	state := "test-state" // Mock state generation
	authURL, err := cfg.GetAuthService().GenerateGoogleAuthURL(state)
	if err != nil {
		cfg.LogHandlerError(
			r.Context(),
			"signin-google",
			"auth_url_generation_failed",
			"Error generating Google auth URL",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to initiate Google signin")
		return
	}

	// Redirect to Google
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandlerGoogleCallback is a test implementation that uses the mocked dependencies
func (cfg *TestHandlersAuthConfig) HandlerGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get parameters from URL
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		cfg.LogHandlerError(
			ctx,
			"callback-google",
			"missing_parameters",
			"Missing state or code parameter",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().HandleGoogleAuth(ctx, code, state)
	if err != nil {
		cfg.handleAuthError(w, r, err, "callback-google", ip, userAgent)
		return
	}

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := ctx // We don't have utils.ContextKeyUserID in test context
	cfg.LogHandlerSuccess(ctxWithUserID, "callback-google", "Google signin success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Google signin successful",
	})
}

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

func TestRealHandlerGoogleSignIn_AuthURLGenerationFailed(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	mockAuthService := &MockAuthService{}
	realAuthConfig := &auth.AuthConfig{}

	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
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

func TestRealHandlerGoogleCallback_MissingParameters(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	mockAuthService := &MockAuthService{}
	realAuthConfig := &auth.AuthConfig{}

	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Auth: realAuthConfig,
		},
		Logger:      mockHandlersConfig,
		authService: mockAuthService,
	}

	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	w := httptest.NewRecorder()

	// Set up mock expectations for the missing parameters error path
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything).Return()

	cfg.HandlerGoogleCallback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing required parameters")
	mockHandlersConfig.AssertExpectations(t)
}
