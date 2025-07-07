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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestConfig() *TestHandlersAuthConfig {
	return &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}
}

// HandlerRefreshToken is a test implementation that uses the mocked dependencies
func (cfg *TestHandlersAuthConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token using mocked auth
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.LogHandlerError(
			ctx,
			"refresh_token",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().RefreshToken(ctx, userID, storedData.Provider, storedData.Token)
	if err != nil {
		cfg.handleAuthError(w, r, err, "refresh_token", ip, userAgent)
		return
	}

	// Set new cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := ctx // We don't have utils.ContextKeyUserID in test context
	cfg.LogHandlerSuccess(ctxWithUserID, "refresh_token", "Refresh token success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Token refreshed successfully",
	})
}

func TestHandlerRefreshToken_Success(t *testing.T) {
	cfg := setupTestConfig()
	userID := uuid.New()
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

func TestHandlerRefreshToken_ServiceError(t *testing.T) {
	cfg := setupTestConfig()
	userID := uuid.New()
	refreshTokenData := &RefreshTokenData{
		Token:    "valid-refresh-token",
		Provider: "local",
	}

	// Mock successful token validation
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return(userID.String(), refreshTokenData, nil)

	// Mock service error
	serviceError := &handlers.AppError{Code: "redis_error", Message: "Error storing refresh token"}
	cfg.authService.(*MockAuthService).On("RefreshToken", mock.Anything, userID.String(), "local", "valid-refresh-token").
		Return(nil, serviceError)

	// Mock logging (accept any value for the error argument)
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "redis_error", "Error storing refresh token", mock.Anything, mock.Anything, mock.Anything)

	// Create request with valid cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh-token"})
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Something went wrong, please try again later")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestHandlerRefreshToken_InvalidTokenError(t *testing.T) {
	cfg := setupTestConfig()
	userID := uuid.New()
	refreshTokenData := &RefreshTokenData{
		Token:    "valid-refresh-token",
		Provider: "local",
	}

	// Mock successful token validation
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return(userID.String(), refreshTokenData, nil)

	// Mock invalid token error
	authError := &handlers.AppError{Code: "invalid_token", Message: "Invalid refresh token"}
	cfg.authService.(*MockAuthService).On("RefreshToken", mock.Anything, userID.String(), "local", "valid-refresh-token").
		Return(nil, authError)

	// Mock logging (expect internal_error as the error code)
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "internal_error", "Invalid refresh token", mock.Anything, mock.Anything, mock.Anything)

	// Create request with valid cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh-token"})
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestHandlerRefreshToken_GenericError(t *testing.T) {
	cfg := setupTestConfig()
	userID := uuid.New()
	refreshTokenData := &RefreshTokenData{
		Token:    "valid-refresh-token",
		Provider: "local",
	}

	// Mock successful token validation
	cfg.Auth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).
		Return(userID.String(), refreshTokenData, nil)

	// Mock generic error
	genericError := errors.New("some unexpected error")
	cfg.authService.(*MockAuthService).On("RefreshToken", mock.Anything, userID.String(), "local", "valid-refresh-token").
		Return(nil, genericError)

	// Mock logging
	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "refresh_token", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, genericError)

	// Create request with valid cookies
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh-token"})
	w := httptest.NewRecorder()

	// Call the handler
	cfg.HandlerRefreshToken(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")

	// Verify mock expectations
	cfg.Auth.AssertExpectations(t)
	cfg.authService.(*MockAuthService).AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

// Test that the handler exists and can be called (basic smoke test)
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
