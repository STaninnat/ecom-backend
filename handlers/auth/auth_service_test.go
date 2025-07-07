package authhandlers

import (
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/stretchr/testify/assert"
)

// TestAuthServiceInterface tests that the AuthService interface is properly defined
func TestAuthServiceInterface(t *testing.T) {
	// This test ensures that the AuthService interface is properly defined
	// and that all required methods are present
	var _ AuthService = (*authServiceImpl)(nil)
}

// TestSignUpParams tests the SignUpParams struct
func TestSignUpParams(t *testing.T) {
	params := SignUpParams{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "testuser", params.Name)
	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, "password123", params.Password)
}

// TestSignInParams tests the SignInParams struct
func TestSignInParams(t *testing.T) {
	params := SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, "password123", params.Password)
}

// TestAuthResult tests the AuthResult struct
func TestAuthResult(t *testing.T) {
	now := time.Now()
	result := &AuthResult{
		UserID:              "user-id",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		AccessTokenExpires:  now,
		RefreshTokenExpires: now,
		IsNewUser:           true,
	}

	assert.Equal(t, "user-id", result.UserID)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, "refresh-token", result.RefreshToken)
	assert.True(t, result.IsNewUser)
}

// TestAuthError tests the AuthError struct
func TestAuthError(t *testing.T) {
	err := &handlers.AppError{
		Code:    "test_error",
		Message: "Test error message",
		Err:     nil,
	}

	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "Test error message", err.Message)
	assert.Nil(t, err.Err)
	assert.Equal(t, "Test error message", err.Error())
}

// TestAuthError_WithInnerError tests AuthError with an inner error
func TestAuthError_WithInnerError(t *testing.T) {
	innerErr := assert.AnError
	err := &handlers.AppError{
		Code:    "test_error",
		Message: "Test error message",
		Err:     innerErr,
	}

	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "Test error message", err.Message)
	assert.Equal(t, innerErr, err.Err)
	assert.Contains(t, err.Error(), "Test error message")
	assert.Contains(t, err.Error(), innerErr.Error())
}

// TestNewAuthService tests the NewAuthService function
func TestNewAuthService(t *testing.T) {
	// This test ensures that NewAuthService returns a valid AuthService
	// The actual implementation would require real dependencies
	// For now, we just test that the function exists and can be called
	assert.NotNil(t, NewAuthService)
}

// TestConstants tests that the constants are properly defined
func TestConstants(t *testing.T) {
	assert.NotZero(t, AccessTokenTTL)
	assert.NotZero(t, RefreshTokenTTL)
	assert.NotZero(t, OAuthStateTTL)
	assert.Equal(t, "local", LocalProvider)
	assert.Equal(t, "user", UserRole)
	assert.Equal(t, "refresh_token:", RefreshTokenKeyPrefix)
	assert.Equal(t, "oauth_state:", OAuthStateKeyPrefix)
	assert.Equal(t, "valid", OAuthStateValid)
}
