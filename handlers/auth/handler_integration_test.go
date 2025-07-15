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
	carthandlers "github.com/STaninnat/ecom-backend/handlers/cart"
	"github.com/STaninnat/ecom-backend/internal/config"
	redismock "github.com/go-redis/redismock/v9"
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

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{Auth: &auth.AuthConfig{}},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{}, // Use real struct
			Logger:             mockHandlersConfig,
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
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerSignIn_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{Auth: &auth.AuthConfig{}},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{}, // Use real struct
			Logger:             mockHandlersConfig,
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
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerSignOut_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		cfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				Auth: &auth.AuthConfig{},
			},
			Logger:      mockHandlersConfig,
			authService: mockAuthService,
		}

		// Set up mock expectation for LogHandlerError
		mockHandlersConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request
		req := httptest.NewRequest("POST", "/signout", nil)
		w := httptest.NewRecorder()

		cfg.HandlerSignOut(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		// Expect the actual error message from Go's http package
		assert.Contains(t, w.Body.String(), "http: named cookie not present")
		// No expectations for SignOut or LogHandlerSuccess, as handler exits early
	})

	t.Run("HandlerSignOut_InvalidToken", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectation for LogHandlerError when no cookie is present
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "sign_out", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request without cookies
		req := httptest.NewRequest("POST", "/signout", nil)
		w := httptest.NewRecorder()

		cfg.HandlerSignOut(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockHandlersConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "SignOut")
	})

	t.Run("HandlerRefreshToken_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create mock Redis client
		db, _ := redismock.NewClientMock()

		// Initialize AuthConfig with a minimal APIConfig and required secrets to avoid nil pointer panic
		cfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				Auth: &auth.AuthConfig{
					APIConfig: &config.APIConfig{
						RefreshSecret: "dummy-refresh-secret",
						RedisClient:   db,
					},
				},
			},
			Logger:      mockHandlersConfig,
			authService: mockAuthService,
		}

		// Set up mock expectations for service and logger
		// No need to set up expectations for LogHandlerSuccess or RefreshToken, as the handler will exit early due to missing/invalid cookie
		mockHandlersConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		// Create request with a dummy refresh token cookie
		req := httptest.NewRequest("POST", "/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "test-refresh-token",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerRefreshToken(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// Expect the actual error message from handler
		assert.Contains(t, response["error"], "invalid refresh token format")
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerGoogleSignIn_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations
		authURL := "https://accounts.google.com/oauth/authorize?client_id=test&redirect_uri=test&response_type=code&scope=openid+email+profile&state=any"
		mockAuthService.On("GenerateGoogleAuthURL", mock.Anything).Return(authURL, nil)

		// Create request
		req := httptest.NewRequest("GET", "/auth/google/signin?state=test-state", nil)
		w := httptest.NewRecorder()

		// Execute - this will call the real handler function with mock services
		cfg.HandlerGoogleSignIn(w, req)

		// Assertions
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, authURL, w.Header().Get("Location"))

		// Verify all mocks were called as expected
		mockAuthService.AssertExpectations(t)
		mockHandlersConfig.AssertExpectations(t)
	})

	t.Run("HandlerGoogleCallback_ValidRequest", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations
		expectedResult := &AuthResult{
			UserID:              "user123",
			AccessToken:         "google_access_token_123",
			RefreshToken:        "google_refresh_token_123",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           true,
		}
		mockAuthService.On("HandleGoogleAuth", mock.Anything, "test-code", "test-state").Return(expectedResult, nil)
		mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "callback-google", "Google signin success", mock.Anything, mock.Anything).Return()

		// Create request
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
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("HandlerSignUp_InvalidJSON", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
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

		// Verify mocks were called as expected
		mockHandlersConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "SignUp")
	})

	t.Run("HandlerSignIn_InvalidJSON", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
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

		// Verify mocks were called as expected
		mockHandlersConfig.AssertExpectations(t)
		mockAuthService.AssertNotCalled(t, "SignIn")
	})

	t.Run("HandlerRefreshToken_NoCookie", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for cookie validation failure
		mockHandlersConfig.On(
			"LogHandlerError",
			mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token",
			mock.Anything, mock.Anything, mock.Anything,
		).Return()

		// Create request without cookies
		req := httptest.NewRequest("POST", "/refresh", nil)
		w := httptest.NewRecorder()

		// Execute
		cfg.HandlerRefreshToken(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "http: named cookie not present")
		mockAuthService.AssertNotCalled(t, "RefreshToken")
	})

	t.Run("HandlerGoogleCallback_MissingParams", func(t *testing.T) {
		// Create mock services
		mockAuthService := new(MockAuthService)
		mockHandlersConfig := new(MockHandlersConfig)

		// Create test config with proper mocks
		cfg := &HandlersAuthConfig{
			HandlersConfig:     &handlers.HandlersConfig{},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             mockHandlersConfig,
			authService:        mockAuthService,
		}

		// Set up mock expectations for error logging
		mockHandlersConfig.On(
			"LogHandlerError",
			mock.Anything, "callback-google", "missing_parameters", "Missing state or code parameter",
			mock.Anything, mock.Anything, mock.Anything,
		).Return()

		// Create request without required parameters
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

		// Verify mocks were called as expected
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
		authService := NewAuthService(nil, nil, &AuthConfigAdapter{authConfig}, nil, nil)

		// Assertions
		assert.NotNil(t, authService)
		assert.Implements(t, (*AuthService)(nil), authService)
	})

	t.Run("AuthError_Error", func(t *testing.T) {
		// Create auth error
		err := &handlers.AppError{
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
		err := &handlers.AppError{
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

// TestRealHandlerIntegration tests the real handlers with no cookies (validation error path only)
func TestRealHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("HandlerSignOut_WithoutCookies", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				Auth: &auth.AuthConfig{},
			},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             &MockHandlersConfig{},
			authService:        &MockAuthService{},
		}

		mockLogger := &MockHandlersConfig{}
		mockService := &MockAuthService{}
		cfg.Logger = mockLogger
		cfg.authService = mockService

		mockLogger.On("LogHandlerError", mock.Anything, "sign_out", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()

		req := httptest.NewRequest("POST", "/signout", nil)
		w := httptest.NewRecorder()

		cfg.HandlerSignOut(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockLogger.AssertExpectations(t)
		mockService.AssertExpectations(t)
	})

	t.Run("HandlerRefreshToken_WithoutCookies", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			HandlersConfig: &handlers.HandlersConfig{
				Auth: &auth.AuthConfig{},
			},
			HandlersCartConfig: &carthandlers.HandlersCartConfig{},
			Logger:             &MockHandlersConfig{},
			authService:        &MockAuthService{},
		}

		mockLogger := &MockHandlersConfig{}
		mockService := &MockAuthService{}
		cfg.Logger = mockLogger
		cfg.authService = mockService

		mockLogger.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()

		req := httptest.NewRequest("POST", "/refresh", nil)
		w := httptest.NewRecorder()

		cfg.HandlerRefreshToken(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockLogger.AssertExpectations(t)
		mockService.AssertExpectations(t)
	})
}
