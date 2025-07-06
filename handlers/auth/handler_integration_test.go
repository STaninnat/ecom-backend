package authhandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerIntegration tests the actual handler functions with mock service implementation
func TestHandlerIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("HandlerSignUp_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Test data
		requestBody := map[string]string{
			"name":     "testuser",
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Set up mock expectations
		expectedResult := &AuthResult{
			UserID:              "user123",
			AccessToken:         "access_token_123",
			RefreshToken:        "refresh_token_123",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           true,
		}
		mockAuthService.On("SignUp", mock.Anything, SignUpParams{
			Name:     "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}).Return(expectedResult, nil)
		mockCartConfig.On("MergeCart", mock.Anything, mock.Anything, "user123").Return()
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "signup-local", "Local signup success", mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerSignUp(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response handlers.HandlerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Signup successful", response.Message)

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockCartConfig.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerSignIn_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Test data
		requestBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Set up mock expectations
		expectedResult := &AuthResult{
			UserID:              "user123",
			AccessToken:         "access_token_123",
			RefreshToken:        "refresh_token_123",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           false,
		}
		mockAuthService.On("SignIn", mock.Anything, SignInParams{
			Email:    "test@example.com",
			Password: "password123",
		}).Return(expectedResult, nil)
		mockCartConfig.On("MergeCart", mock.Anything, mock.Anything, "user123").Return()
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "signin-local", "Local signin success", mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerSignIn(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.HandlerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Signin successful", response.Message)

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockCartConfig.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerSignOut_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for cookie validation
		mockAuthConfig.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return("user123", &RefreshTokenData{
			Token:    "refresh_token_123",
			Provider: "local",
		}, nil)
		mockAuthService.On("SignOut", mock.Anything, "user123", "local").Return(nil)
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "sign_out", "Sign out success", mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signout", nil)
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerSignOut(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.HandlerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Sign out successful", response.Message)

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerRefreshToken_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for cookie validation
		mockAuthConfig.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return("user123", &RefreshTokenData{
			Token:    "refresh_token_123",
			Provider: "local",
		}, nil)

		// Set up mock expectations for token refresh
		expectedResult := &AuthResult{
			UserID:              "user123",
			AccessToken:         "new_access_token_123",
			RefreshToken:        "new_refresh_token_123",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           false,
		}
		mockAuthService.On("RefreshToken", mock.Anything, "user123", "local", "refresh_token_123").Return(expectedResult, nil)
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "refresh_token", "Refresh token success", mock.Anything, mock.Anything).Return()

		// Create request with refresh token cookie
		req := httptest.NewRequest("POST", "/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "test-refresh-token",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false, // For testing
			SameSite: http.SameSiteLaxMode,
		})
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerRefreshToken(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.HandlerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Token refreshed successfully", response.Message)

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerGoogleSignIn_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations
		mockAuthService.On("GenerateGoogleAuthURL", mock.Anything).Return("https://accounts.google.com/oauth2/auth?state=test-state", nil)

		// Create request
		req := httptest.NewRequest("GET", "/auth/google", nil)
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerGoogleSignIn(w, req)

		// Assertions - Google SignIn should redirect (302)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "https://accounts.google.com/oauth2/auth?state=test-state", w.Header().Get("Location"))

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
	})

	t.Run("HandlerGoogleCallback_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations
		expectedResult := &AuthResult{
			UserID:              "user123",
			AccessToken:         "access_token_123",
			RefreshToken:        "refresh_token_123",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           true,
		}
		mockAuthService.On("HandleGoogleAuth", mock.Anything, "test-code", "test-state").Return(expectedResult, nil)
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "callback-google", "Google signin success", mock.Anything, mock.Anything).Return()

		// Create request with valid state and code
		req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code&state=test-state", nil)
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerGoogleCallback(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response handlers.HandlerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Google signin successful", response.Message)

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})
}

// TestHandlerErrorScenarios tests error handling in handlers
func TestHandlerErrorScenarios(t *testing.T) {
	t.Run("HandlerSignUp_InvalidJSON", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Test data - invalid JSON
		invalidJSON := `{"name": "testuser", "email": "test@example.com", "password": "password123"`

		// Set up mock expectations for error logging
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "invalid_request", "Invalid signup payload", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute
		cfg.HandlerSignUp(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request payload", response["error"])

		// Verify mock was called
		mockHandlersConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "SignUp")
	})

	t.Run("HandlerSignIn_InvalidJSON", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Test data - invalid JSON
		invalidJSON := `{"email": "test@example.com", "password": "password123"`

		// Set up mock expectations for error logging
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_request", "Invalid signin payload", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signin", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute
		cfg.HandlerSignIn(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request payload", response["error"])
	})

	t.Run("HandlerRefreshToken_NoCookie", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for cookie validation failure
		mockAuthConfig.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return("", (*RefreshTokenData)(nil), assert.AnError)
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request without refresh token cookie
		req := httptest.NewRequest("POST", "/refresh", nil)
		w := httptest.NewRecorder()

		// Execute
		cfg.HandlerRefreshToken(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "assert.AnError general error for testing", response["error"])

		// Verify mock was called
		mockAuthConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "RefreshToken")
	})

	t.Run("HandlerGoogleCallback_MissingParams", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)
		mockCartConfig := new(MockCartConfig)
		mockAuthConfig := new(mockAuthConfig)

		// Create test config with proper mocks
		cfg := &TestHandlersAuthConfig{
			MockHandlersConfig: mockHandlersConfig,
			MockCartConfig:     mockCartConfig,
			Auth:               mockAuthConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for error logging
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request without code and state parameters
		req := httptest.NewRequest("GET", "/auth/google/callback", nil)
		w := httptest.NewRecorder()

		// Execute
		cfg.HandlerGoogleCallback(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Missing required parameters", response["error"])

		// Verify mock was called
		mockHandlersConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "HandleGoogleAuth")
	})
}

// TestAuthWrapperIntegration tests the auth wrapper functions
func TestAuthWrapperIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InitAuthService_MissingDependencies", func(t *testing.T) {
		// Create test config with minimal setup
		apiCfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{},
		}

		// Execute
		err := apiCfg.InitAuthService()

		// Assertions - should fail due to missing dependencies, but gracefully
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("InitAuthService_Success", func(t *testing.T) {
		// Create test config with minimal setup - this will fail gracefully
		apiCfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				APIConfig: &config.APIConfig{},
			},
		}

		// Execute - this should fail due to missing dependencies but not panic
		err := apiCfg.InitAuthService()

		// Assertions - should fail gracefully with proper error message
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetAuthService_Success", func(t *testing.T) {
		// Create test config
		apiCfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				APIConfig: &config.APIConfig{},
			},
		}

		// Execute - this should work even with nil dependencies
		authService := apiCfg.GetAuthService()

		// Assertions - should return a service instance even with nil dependencies
		assert.NotNil(t, authService)
		assert.Implements(t, (*AuthService)(nil), authService)
	})
}

// TestAuthServiceMethods tests the actual auth service methods
func TestAuthServiceMethods(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("NewAuthService_Success", func(t *testing.T) {
		// Create test dependencies
		authConfig := &auth.AuthConfig{
			APIConfig: &config.APIConfig{
				JWTSecret:     "test-secret-key-for-testing-only-32-chars-long-enough",
				RefreshSecret: "test-refresh-secret-key-for-testing-only-32-chars-long-enough",
			},
		}

		// Execute
		authService := NewAuthService(nil, nil, authConfig, nil, nil)

		// Assertions
		assert.NotNil(t, authService)
		assert.Implements(t, (*AuthService)(nil), authService)
	})

	t.Run("AuthError_Error", func(t *testing.T) {
		// Create auth error
		err := &AuthError{
			Code:    "test_error",
			Message: "Test error message",
		}

		// Execute
		errorMsg := err.Error()

		// Assertions
		assert.Equal(t, "Test error message", errorMsg)
	})

	t.Run("AuthError_Unwrap", func(t *testing.T) {
		// Create auth error with inner error
		innerErr := assert.AnError
		err := &AuthError{
			Code:    "test_error",
			Message: "Test error message",
			Err:     innerErr,
		}

		// Execute
		unwrappedErr := err.Unwrap()

		// Assertions
		assert.Equal(t, innerErr, unwrappedErr)
	})
}
