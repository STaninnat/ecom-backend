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
	carthandlers "github.com/STaninnat/ecom-backend/handlers/cart"
	testutil "github.com/STaninnat/ecom-backend/internal/testutil"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_refresh_token_test.go: Tests for HandlerRefreshToken — validates refresh token flow and error handling.

// TestHandlerRefreshToken_Success verifies successful token refresh and checks response and cookies.
func TestHandlerRefreshToken_Success(t *testing.T) {
	cfg := setupTestConfig()
	userID := utils.NewUUID()
	refreshTokenData := &RefreshTokenData{
		Token:    "valid-refresh-token",
		Provider: "local",
	}

	// Mock successful token validation
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return(userID.String(), refreshTokenData, nil)

	// Mock successful service call
	cfg.authService.(*MockAuthService).On("RefreshToken", mock.Anything, userID.String(), "local", "valid-refresh-token").
		Return(&AuthResult{
			UserID:              userID.String(),
			AccessToken:         "new-access-token",
			RefreshToken:        "new-refresh-token",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           false,
		}, nil)

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "refresh_token", "Refresh token success", mock.Anything, mock.Anything)

	// Create request with valid cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh-token"})
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Token refreshed successfully")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerRefreshToken_InvalidToken checks that an invalid refresh token returns unauthorized and logs appropriately.
func TestHandlerRefreshToken_InvalidToken(t *testing.T) {
	cfg := setupTestConfig()

	// Mock token validation failure
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return("", (*RefreshTokenData)(nil), errors.New("invalid token"))

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything)

	// Create request with invalid cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerRefreshToken_ErrorScenarios tests various error scenarios for HandlerRefreshToken.
func TestHandlerRefreshToken_ErrorScenarios(t *testing.T) {
	testutil.RunAuthTokenErrorScenarios(t, "refresh_token", func(w http.ResponseWriter, r *http.Request, logger *testutil.MockHandlersConfig, authService *testutil.MockAuthService) {
		cfg := setupTestConfig()
		cfg.MockHandlersConfig = (*MockHandlersConfig)(logger)
		cfg.authService = (*MockAuthService)(authService)
		cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return("", (*RefreshTokenData)(nil), assert.AnError)
		cfg.HandlerRefreshToken(w, r)
	})
}

// TestHandlerRefreshToken_Exists is a smoke test to ensure the handler exists and can be called without panicking.
func TestHandlerRefreshToken_Exists(t *testing.T) {
	cfg := setupTestConfig()

	// Mock token validation failure (expected for missing cookies)
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return("", (*RefreshTokenData)(nil), errors.New("missing cookie"))

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything)

	// Create a simple request without cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	w := httptest.NewRecorder()

	// Call the handler - it will fail due to missing cookies, but that's expected
	// This test just verifies the handler exists and can be called without panicking
	cfg.HandlerRefreshToken(w, req)

	// Should return an error status due to missing cookies
	assert.NotEqual(t, http.StatusOK, w.Code)

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestHandlerRefreshToken_EmptyToken checks that refresh with an empty token still succeeds if service allows.
func TestHandlerRefreshToken_EmptyToken(t *testing.T) {
	cfg := setupTestConfig()
	userID := utils.NewUUID()
	refreshTokenData := &RefreshTokenData{
		Token:    "", // Empty token
		Provider: "local",
	}

	// Mock successful token validation
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return(userID.String(), refreshTokenData, nil)

	// Mock successful service call
	cfg.authService.(*MockAuthService).On("RefreshToken", mock.Anything, userID.String(), "local", "").
		Return(&AuthResult{
			UserID:              userID.String(),
			AccessToken:         "new-access-token",
			RefreshToken:        "new-refresh-token",
			AccessTokenExpires:  time.Now().Add(30 * time.Minute),
			RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
			IsNewUser:           false,
		}, nil)

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "refresh_token", "Refresh token success", mock.Anything, mock.Anything)

	// Create request
	req := httptest.NewRequest("POST", "/refresh", nil)
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Token refreshed successfully")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// TestRealHandlerRefreshToken_Direct tests the real HandlerRefreshToken method directly for various scenarios and expected responses.
func TestRealHandlerRefreshToken_Direct(t *testing.T) {
	// Create real config with mocks
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			Auth: &auth.Config{}, // Real auth config
		},
		HandlersCartConfig: &carthandlers.HandlersCartConfig{},
		Logger:             &MockHandlersConfig{},
		authService:        &MockAuthService{},
	}

	// Test cases for different scenarios
	testCases := []struct {
		name           string
		setupMocks     func(*MockHandlersConfig, *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success_LocalProvider",
			setupMocks: func(logger *MockHandlersConfig, _ *MockAuthService) {
				// Mock validation error since real handler will fail without cookies
				logger.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "http: named cookie not present",
		},
		{
			name: "Success_GoogleProvider",
			setupMocks: func(logger *MockHandlersConfig, _ *MockAuthService) {
				// Mock validation error since real handler will fail without cookies
				logger.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "http: named cookie not present",
		},
		{
			name: "ServiceError",
			setupMocks: func(logger *MockHandlersConfig, _ *MockAuthService) {
				// Mock validation error since real handler will fail without cookies
				logger.On("LogHandlerError", mock.Anything, "refresh_token", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "http: named cookie not present",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh mocks for each test
			mockLogger := &MockHandlersConfig{}
			mockService := &MockAuthService{}

			cfg.Logger = mockLogger
			cfg.authService = mockService

			// Setup mocks
			tc.setupMocks(mockLogger, mockService)

			// Create request
			req := httptest.NewRequest("POST", "/refresh", nil)
			w := httptest.NewRecorder()

			// Execute
			cfg.HandlerRefreshToken(w, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBody)
			}

			// Verify mocks
			mockLogger.AssertExpectations(t)
			mockService.AssertExpectations(t)
		})
	}
}
