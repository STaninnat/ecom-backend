package handlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
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

// Edge case tests for handlers_embedded.go

func TestNewHandlerConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name                   string
		authService            AuthService
		userService            UserService
		loggerService          LoggerService
		requestMetadataService RequestMetadataService
		jwtSecret              string
		refreshSecret          string
		issuer                 string
		audience               string
		oauth                  *OAuthConfig
		customTokenSource      func(ctx context.Context, refreshToken string) oauth2.TokenSource
		expectNil              bool
	}{
		{
			name:                   "all nil parameters",
			authService:            nil,
			userService:            nil,
			loggerService:          nil,
			requestMetadataService: nil,
			jwtSecret:              "",
			refreshSecret:          "",
			issuer:                 "",
			audience:               "",
			oauth:                  nil,
			customTokenSource:      nil,
			expectNil:              false, // Should still create config
		},
		{
			name:                   "empty strings",
			authService:            &MockAuthService{},
			userService:            &MockUserService{},
			loggerService:          &MockLoggerService{},
			requestMetadataService: &MockRequestMetadataService{},
			jwtSecret:              "",
			refreshSecret:          "",
			issuer:                 "",
			audience:               "",
			oauth:                  &OAuthConfig{},
			customTokenSource:      func(ctx context.Context, refreshToken string) oauth2.TokenSource { return nil },
			expectNil:              false,
		},
		{
			name:                   "nil custom token source",
			authService:            &MockAuthService{},
			userService:            &MockUserService{},
			loggerService:          &MockLoggerService{},
			requestMetadataService: &MockRequestMetadataService{},
			jwtSecret:              "secret",
			refreshSecret:          "refresh",
			issuer:                 "issuer",
			audience:               "audience",
			oauth:                  &OAuthConfig{},
			customTokenSource:      nil,
			expectNil:              false,
		},
		{
			name:                   "nil oauth config",
			authService:            &MockAuthService{},
			userService:            &MockUserService{},
			loggerService:          &MockLoggerService{},
			requestMetadataService: &MockRequestMetadataService{},
			jwtSecret:              "secret",
			refreshSecret:          "refresh",
			issuer:                 "issuer",
			audience:               "audience",
			oauth:                  nil,
			customTokenSource:      func(ctx context.Context, refreshToken string) oauth2.TokenSource { return nil },
			expectNil:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewHandlerConfig(
				tt.authService,
				tt.userService,
				tt.loggerService,
				tt.requestMetadataService,
				tt.jwtSecret,
				tt.refreshSecret,
				tt.issuer,
				tt.audience,
				tt.oauth,
				tt.customTokenSource,
			)

			if tt.expectNil {
				assert.Nil(t, cfg)
			} else {
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.authService, cfg.AuthService)
				assert.Equal(t, tt.userService, cfg.UserService)
				assert.Equal(t, tt.loggerService, cfg.LoggerService)
				assert.Equal(t, tt.requestMetadataService, cfg.RequestMetadataService)
				assert.Equal(t, tt.jwtSecret, cfg.JWTSecret)
				assert.Equal(t, tt.refreshSecret, cfg.RefreshSecret)
				assert.Equal(t, tt.issuer, cfg.Issuer)
				assert.Equal(t, tt.audience, cfg.Audience)
				assert.Equal(t, tt.oauth, cfg.OAuth)
				assert.Equal(t, tt.customTokenSource != nil, cfg.CustomTokenSource != nil)
			}
		})
	}
}

func TestHandlersConfig_ValidateConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *HandlersConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "completely nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name: "all fields nil",
			config: &HandlersConfig{
				Logger:    nil,
				Auth:      nil,
				APIConfig: nil,
				OAuth:     nil,
			},
			expectError: true,
			errorMsg:    "logger is required",
		},
		{
			name: "only logger set",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      nil,
				APIConfig: nil,
			},
			expectError: true,
			errorMsg:    "auth configuration is required",
		},
		{
			name: "logger and auth set, api config nil",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      &auth.AuthConfig{},
				APIConfig: nil,
			},
			expectError: true,
			errorMsg:    "API configuration is required",
		},
		{
			name: "logger and api config set, auth nil",
			config: &HandlersConfig{
				Logger:    logrus.New(),
				Auth:      nil,
				APIConfig: &config.APIConfig{},
			},
			expectError: true,
			errorMsg:    "auth configuration is required",
		},
		{
			name: "auth and api config set, logger nil",
			config: &HandlersConfig{
				Logger:    nil,
				Auth:      &auth.AuthConfig{},
				APIConfig: &config.APIConfig{},
			},
			expectError: true,
			errorMsg:    "logger is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				// Test nil config panic
				assert.Panics(t, func() {
					tt.config.ValidateConfig()
				})
			} else {
				err := tt.config.ValidateConfig()

				if tt.expectError {
					assert.Error(t, err)
					assert.Equal(t, tt.errorMsg, err.Error())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestErrInvalidConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
	}{
		{
			name:     "empty string",
			errorMsg: "",
		},
		{
			name:     "special characters",
			errorMsg: "error with special chars: !@#$%^&*()",
		},
		{
			name:     "unicode characters",
			errorMsg: "error with unicode: 你好世界",
		},
		{
			name:     "very long message",
			errorMsg: "this is a very long error message that contains many characters and should still work properly even when it's extremely long and contains lots of text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ErrInvalidConfig(tt.errorMsg)
			assert.Error(t, err)
			assert.Equal(t, tt.errorMsg, err.Error())
		})
	}
}

func TestHandlerResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "empty message",
			message: "",
		},
		{
			name:    "unicode message",
			message: "你好世界",
		},
		{
			name:    "special characters",
			message: "message with !@#$%^&*()",
		},
		{
			name:    "very long message",
			message: "this is a very long message that contains many characters and should still work properly even when it's extremely long and contains lots of text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := HandlerResponse{
				Message: tt.message,
			}
			assert.Equal(t, tt.message, response.Message)
		})
	}
}

func TestOAuthConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		oauth *OAuthConfig
	}{
		{
			name:  "nil oauth config",
			oauth: nil,
		},
		{
			name: "oauth config with nil google",
			oauth: &OAuthConfig{
				Google: nil,
			},
		},
		{
			name: "oauth config with empty google config",
			oauth: &OAuthConfig{
				Google: &oauth2.Config{
					ClientID:     "",
					ClientSecret: "",
				},
			},
		},
		{
			name: "oauth config with special characters",
			oauth: &OAuthConfig{
				Google: &oauth2.Config{
					ClientID:     "client!@#$%^&*()",
					ClientSecret: "secret!@#$%^&*()",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.oauth == nil {
				assert.Nil(t, tt.oauth)
			} else {
				assert.NotNil(t, tt.oauth)
				if tt.oauth.Google == nil {
					assert.Nil(t, tt.oauth.Google)
				} else {
					assert.NotNil(t, tt.oauth.Google)
					assert.Equal(t, tt.oauth.Google.ClientID, tt.oauth.Google.ClientID)
					assert.Equal(t, tt.oauth.Google.ClientSecret, tt.oauth.Google.ClientSecret)
				}
			}
		})
	}
}

func TestClaims_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "empty user ID",
			userID: "",
		},
		{
			name:   "unicode user ID",
			userID: "用户123",
		},
		{
			name:   "special characters in user ID",
			userID: "user!@#$%^&*()",
		},
		{
			name:   "very long user ID",
			userID: "this_is_a_very_long_user_id_that_contains_many_characters_and_should_still_work_properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				UserID: tt.userID,
			}
			assert.Equal(t, tt.userID, claims.UserID)
		})
	}
}

func TestRefreshTokenData_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		provider string
	}{
		{
			name:     "empty token and provider",
			token:    "",
			provider: "",
		},
		{
			name:     "unicode values",
			token:    "token你好世界",
			provider: "provider你好世界",
		},
		{
			name:     "special characters",
			token:    "token!@#$%^&*()",
			provider: "provider!@#$%^&*()",
		},
		{
			name:     "very long values",
			token:    "this_is_a_very_long_token_that_contains_many_characters_and_should_still_work_properly",
			provider: "this_is_a_very_long_provider_name_that_contains_many_characters_and_should_still_work_properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &RefreshTokenData{
				Token:    tt.token,
				Provider: tt.provider,
			}
			assert.Equal(t, tt.token, data.Token)
			assert.Equal(t, tt.provider, data.Provider)
		})
	}
}

func TestTokenTTLConstants_EdgeCases(t *testing.T) {
	// Test that constants are positive values
	assert.True(t, AccessTokenTTL > 0)
	assert.True(t, RefreshTokenTTL > 0)

	// Test that refresh token TTL is longer than access token TTL
	assert.True(t, RefreshTokenTTL > AccessTokenTTL)

	// Test that constants are reasonable values (not too short or too long)
	assert.True(t, AccessTokenTTL >= 5*time.Minute)  // At least 5 minutes
	assert.True(t, AccessTokenTTL <= 60*time.Minute) // At most 1 hour

	assert.True(t, RefreshTokenTTL >= 24*time.Hour)    // At least 1 day
	assert.True(t, RefreshTokenTTL <= 30*24*time.Hour) // At most 30 days
}

func TestHandlerTypes_EdgeCases(t *testing.T) {
	// Test that handler types can be assigned nil
	var authHandler AuthHandler
	var optionalHandler OptionalHandler

	// These should compile without errors
	assert.Nil(t, authHandler)
	assert.Nil(t, optionalHandler)

	// Test that handler types can be assigned functions
	authHandler = func(w http.ResponseWriter, r *http.Request, user database.User) {
		// Empty function
	}
	optionalHandler = func(w http.ResponseWriter, r *http.Request, user *database.User) {
		// Empty function
	}

	assert.NotNil(t, authHandler)
	assert.NotNil(t, optionalHandler)
}
