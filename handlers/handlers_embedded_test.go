package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestNewHandlerConfig(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	jwtSecret := "test-jwt-secret"
	refreshSecret := "test-refresh-secret"
	issuer := "test-issuer"
	audience := "test-audience"
	oauth := &OAuthConfig{
		Google: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}

	customTokenSource := func(ctx context.Context, refreshToken string) oauth2.TokenSource {
		return nil
	}

	cfg := NewHandlerConfig(
		mockAuth,
		mockUser,
		mockLogger,
		mockRequestMetadata,
		jwtSecret,
		refreshSecret,
		issuer,
		audience,
		oauth,
		customTokenSource,
	)

	assert.NotNil(t, cfg)
	assert.Equal(t, mockAuth, cfg.AuthService)
	assert.Equal(t, mockUser, cfg.UserService)
	assert.Equal(t, mockLogger, cfg.LoggerService)
	assert.Equal(t, mockRequestMetadata, cfg.RequestMetadataService)
	assert.Equal(t, jwtSecret, cfg.JWTSecret)
	assert.Equal(t, refreshSecret, cfg.RefreshSecret)
	assert.Equal(t, issuer, cfg.Issuer)
	assert.Equal(t, audience, cfg.Audience)
	assert.Equal(t, oauth, cfg.OAuth)
	assert.NotNil(t, cfg.CustomTokenSource)
}

func TestHandlersConfig_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *HandlersConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      &auth.AuthConfig{},
				APIConfig: &config.APIConfig{},
			},
			expectError: false,
		},
		{
			name: "nil logger",
			config: &HandlersConfig{
				Logger:    nil,
				Auth:      &auth.AuthConfig{},
				APIConfig: &config.APIConfig{},
			},
			expectError: true,
			errorMsg:    "logger is required",
		},
		{
			name: "nil auth",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      nil,
				APIConfig: &config.APIConfig{},
			},
			expectError: true,
			errorMsg:    "auth configuration is required",
		},
		{
			name: "nil API config",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      &auth.AuthConfig{},
				APIConfig: nil,
			},
			expectError: true,
			errorMsg:    "API configuration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrInvalidConfig(t *testing.T) {
	errorMsg := "test error message"
	err := ErrInvalidConfig(errorMsg)

	assert.Error(t, err)
	assert.Equal(t, errorMsg, err.Error())
}

func TestTokenTTLConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, 30*time.Minute, AccessTokenTTL)
	assert.Equal(t, 7*24*time.Hour, RefreshTokenTTL)
}

func TestHandlerResponse(t *testing.T) {
	message := "test message"
	response := HandlerResponse{
		Message: message,
	}

	assert.Equal(t, message, response.Message)
}

func TestOAuthConfig(t *testing.T) {
	oauth := &OAuthConfig{
		Google: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}

	assert.NotNil(t, oauth)
	assert.NotNil(t, oauth.Google)
	assert.Equal(t, "test-client-id", oauth.Google.ClientID)
	assert.Equal(t, "test-client-secret", oauth.Google.ClientSecret)
}

func TestHandlerTypes(t *testing.T) {
	// Test that handler types are properly defined
	var authHandler AuthHandler
	var optionalHandler OptionalHandler

	// These should compile without errors
	assert.Nil(t, authHandler)
	assert.Nil(t, optionalHandler)
}

func TestClaims(t *testing.T) {
	userID := "test-user-id"
	claims := &Claims{
		UserID: userID,
	}

	assert.Equal(t, userID, claims.UserID)
}

func TestRefreshTokenData(t *testing.T) {
	token := "test-token"
	provider := "test-provider"
	data := &RefreshTokenData{
		Token:    token,
		Provider: provider,
	}

	assert.Equal(t, token, data.Token)
	assert.Equal(t, provider, data.Provider)
}

// Note: SetupHandlersConfig is not tested here because it requires
// actual configuration files and database connections.
// In a real test environment, you would use test containers or mocks
// to test this function.
