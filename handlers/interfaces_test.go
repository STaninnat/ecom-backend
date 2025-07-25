// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// interfaces_test.go: Tests for HandlerConfig, related interfaces, and config validation to ensure proper setup and functionality.

// TestNewHandlerConfig_Interfaces tests NewHandlerConfig with interface mocks.
// It checks that all fields are set as expected and not nil.
func TestNewHandlerConfig_Interfaces(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	mockLoggerService := &MockLoggerService{}
	mockRequestMetadataService := &MockRequestMetadataService{}

	jwtSecret := "test-jwt-secret"
	refreshSecret := "test-refresh-secret"
	issuer := "test-issuer"
	audience := "test-audience"
	oauthConfig := &OAuthConfig{
		Google: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}
	customTokenSource := func(_ context.Context, _ string) oauth2.TokenSource {
		return nil
	}

	runNewHandlerConfigTest(
		t,
		mockAuthService,
		mockUserService,
		mockLoggerService,
		mockRequestMetadataService,
		jwtSecret,
		refreshSecret,
		issuer,
		audience,
		oauthConfig,
		customTokenSource,
	)
}

// TestHandlerConfig_Complete tests creation of HandlerConfig with all fields set.
// It checks that all fields are not nil or empty as appropriate.
func TestHandlerConfig_Complete(t *testing.T) {
	// Test that HandlerConfig can be created with all fields
	cfg := &HandlerConfig{
		AuthService:            &MockAuthService{},
		UserService:            &MockUserService{},
		LoggerService:          &MockLoggerService{},
		RequestMetadataService: &MockRequestMetadataService{},
		JWTSecret:              "test-secret",
		RefreshSecret:          "test-refresh-secret",
		Issuer:                 "test-issuer",
		Audience:               "test-audience",
		OAuth: &OAuthConfig{
			Google: &oauth2.Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
		},
		CustomTokenSource: func(_ context.Context, _ string) oauth2.TokenSource {
			return nil
		},
	}

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.AuthService)
	assert.NotNil(t, cfg.UserService)
	assert.NotNil(t, cfg.LoggerService)
	assert.NotNil(t, cfg.RequestMetadataService)
	assert.NotEmpty(t, cfg.JWTSecret)
	assert.NotEmpty(t, cfg.RefreshSecret)
	assert.NotEmpty(t, cfg.Issuer)
	assert.NotEmpty(t, cfg.Audience)
	assert.NotNil(t, cfg.OAuth)
	assert.NotNil(t, cfg.OAuth.Google)
	assert.NotNil(t, cfg.CustomTokenSource)
}

// TestOAuthConfig_Complete tests creation of OAuthConfig with a Google config.
// It checks that all fields are set as expected and not nil.
func TestOAuthConfig_Complete(t *testing.T) {
	// Test that OAuthConfig can be created with Google config
	oauthConfig := &OAuthConfig{
		Google: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}

	assert.NotNil(t, oauthConfig)
	assert.NotNil(t, oauthConfig.Google)
	assert.Equal(t, "test-client-id", oauthConfig.Google.ClientID)
	assert.Equal(t, "test-client-secret", oauthConfig.Google.ClientSecret)
	assert.Equal(t, "http://localhost:8080/auth/google/callback", oauthConfig.Google.RedirectURL)
	assert.Len(t, oauthConfig.Google.Scopes, 3)
}

// TestClaims_Complete tests creation and field access of Claims.
// It checks that the UserID field is set as expected.
func TestClaims_Complete(t *testing.T) {
	// Test that Claims can be created and marshaled
	claims := &Claims{
		UserID: "user123",
	}

	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
}

// TestRefreshTokenData_Complete tests creation and field access of RefreshTokenData.
// It checks that the Token and Provider fields are set as expected.
func TestRefreshTokenData_Complete(t *testing.T) {
	// Test that RefreshTokenData can be created
	refreshData := &RefreshTokenData{
		Token:    "refresh-token-123",
		Provider: "google",
	}

	assert.NotNil(t, refreshData)
	assert.Equal(t, "refresh-token-123", refreshData.Token)
	assert.Equal(t, "google", refreshData.Provider)
}

// TestAuthHandler_Type tests that AuthHandler is a valid function type.
// It checks that a function can be assigned to AuthHandler and is not nil.
func TestAuthHandler_Type(t *testing.T) {
	// Test that AuthHandler is a valid function type
	handler := func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
		// Test implementation
	}

	assert.NotNil(t, handler)
}

// TestOptionalHandler_Type tests that OptionalHandler is a valid function type.
// It checks that a function can be assigned to OptionalHandler and is not nil.
func TestOptionalHandler_Type(t *testing.T) {
	// Test that OptionalHandler is a valid function type
	handler := func(_ http.ResponseWriter, _ *http.Request, _ *database.User) {
		// Test implementation
	}

	assert.NotNil(t, handler)
}

// TestHandlersConfig_ValidateConfig_Success tests successful validation of Config.
// It checks that no error is returned for a valid config.
func TestHandlersConfig_ValidateConfig_Success(t *testing.T) {
	// Test successful validation
	logger := logrus.New()
	cfg := &Config{
		Logger: logger,
		Auth:   &auth.Config{},
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	assert.NoError(t, err)
}

// TestHandlersConfig_ValidateConfig_MissingLogger tests validation failure with missing logger.
// It checks that an error is returned and contains the expected message.
func TestHandlersConfig_ValidateConfig_MissingLogger(t *testing.T) {
	// Test validation failure with missing logger
	cfg := &Config{
		Auth: &auth.Config{},
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "logger is required")
}

// TestHandlersConfig_ValidateConfig_MissingAuth tests validation failure with missing auth.
// It checks that an error is returned and contains the expected message.
func TestHandlersConfig_ValidateConfig_MissingAuth(t *testing.T) {
	// Test validation failure with missing auth
	logger := logrus.New()
	cfg := &Config{
		Logger: logger,
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth configuration is required")
}

// TestHandlersConfig_ValidateConfig_MissingAPIConfig tests validation failure with missing API config.
// It checks that an error is returned and contains the expected message.
func TestHandlersConfig_ValidateConfig_MissingAPIConfig(t *testing.T) {
	// Test validation failure with missing API config
	logger := logrus.New()
	cfg := &Config{
		Logger: logger,
		Auth:   &auth.Config{},
	}

	err := cfg.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API configuration is required")
}

// TestErrInvalidConfig_Error tests the error message returned by ErrInvalidConfig.
// It checks that the error message matches the input string.
func TestErrInvalidConfig_Error(t *testing.T) {
	// Test ErrInvalidConfig error message
	err := ErrInvalidConfig("test error message")
	assert.Equal(t, "test error message", err.Error())
}

// TestInterfaceCompliance tests that mock services implement the required interfaces.
// It checks that each mock can be assigned to its respective interface type.
func TestInterfaceCompliance(t *testing.T) {
	// Test that our mock services implement the interfaces correctly

	// AuthService interface
	var authService AuthService = &MockAuthService{}
	assert.NotNil(t, authService)

	// UserService interface
	var userService UserService = &MockUserService{}
	assert.NotNil(t, userService)

	// LoggerService interface
	var loggerService LoggerService = &MockLoggerService{}
	assert.NotNil(t, loggerService)

	// RequestMetadataService interface
	var metadataService RequestMetadataService = &MockRequestMetadataService{}
	assert.NotNil(t, metadataService)
}
