package authhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Extend TestHandlersAuthConfig to include Auth field and HandlerSignOut method
func (cfg *TestHandlersAuthConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.LogHandlerError(
			ctx,
			"sign_out",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	err = cfg.GetAuthService().SignOut(ctx, userID, storedData.Provider)
	if err != nil {
		cfg.handleAuthError(w, r, err, "sign_out", ip, userAgent)
		return
	}

	// Clear cookies
	timeNow := time.Now().UTC()
	expiredTime := timeNow.Add(-1 * time.Hour)
	auth.SetTokensAsCookies(w, "", "", expiredTime, expiredTime)

	// Handle Google revoke if needed
	if storedData.Provider == "google" {
		googleRevokeURL := "https://accounts.google.com/o/oauth2/revoke?token=" + storedData.Token
		http.Redirect(w, r, googleRevokeURL, http.StatusFound)
		return
	}

	// Log success and respond
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.LogHandlerSuccess(ctxWithUserID, "sign_out", "Sign out success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Sign out successful",
	})
}

func TestHandlerSignOut_Success(t *testing.T) {
	// Setup
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}

	// Create test data
	userID := "test-user-id"
	storedData := &RefreshTokenData{
		Token:    "test-refresh-token",
		Provider: "local",
	}

	// Create request
	req := httptest.NewRequest("POST", "/signout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockAuth := cfg.Auth
	mockAuth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return(userID, storedData, nil)

	mockService := cfg.authService.(*MockAuthService)
	mockService.On("SignOut", mock.Anything, userID, storedData.Provider).Return(nil)

	cfg.MockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "sign_out", "Sign out success", mock.Anything, mock.Anything).Return()

	// Execute
	cfg.HandlerSignOut(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Sign out successful", response.Message)

	// Verify mock calls
	mockAuth.AssertExpectations(t)
	mockService.AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignOut_InvalidToken(t *testing.T) {
	// Setup
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}

	// Create request
	req := httptest.NewRequest("POST", "/signout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockAuth := cfg.Auth
	mockAuth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return("", (*RefreshTokenData)(nil), errors.New("invalid token"))

	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "sign_out", "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()

	// Execute
	cfg.HandlerSignOut(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid token", response["error"])

	// Verify mock calls
	mockAuth.AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignOut_SignOutFailure(t *testing.T) {
	// Setup
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}

	// Create test data
	userID := "test-user-id"
	storedData := &RefreshTokenData{
		Token:    "test-refresh-token",
		Provider: "local",
	}

	// Create request
	req := httptest.NewRequest("POST", "/signout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockAuth := cfg.Auth
	mockAuth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return(userID, storedData, nil)

	mockService := cfg.authService.(*MockAuthService)
	mockService.On("SignOut", mock.Anything, userID, storedData.Provider).Return(errors.New("signout failed"))

	cfg.MockHandlersConfig.On("LogHandlerError", mock.Anything, "sign_out", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, mock.Anything).Return()

	// Execute
	cfg.HandlerSignOut(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	// Verify mock calls
	mockAuth.AssertExpectations(t)
	mockService.AssertExpectations(t)
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignOut_GoogleProvider(t *testing.T) {
	// Setup
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}

	// Create test data
	userID := "test-user-id"
	storedData := &RefreshTokenData{
		Token:    "google-refresh-token",
		Provider: "google",
	}

	// Create request
	req := httptest.NewRequest("POST", "/signout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockAuth := cfg.Auth
	mockAuth.On("ValidateCookieRefreshTokenData", mock.Anything, mock.Anything).Return(userID, storedData, nil)

	mockService := cfg.authService.(*MockAuthService)
	mockService.On("SignOut", mock.Anything, userID, storedData.Provider).Return(nil)

	// Execute
	cfg.HandlerSignOut(w, req)

	// Assertions
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "https://accounts.google.com/o/oauth2/revoke?token=google-refresh-token")

	// Verify mock calls
	mockAuth.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
