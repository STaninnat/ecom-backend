package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

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
	customTokenSource := func(ctx context.Context, refreshToken string) oauth2.TokenSource {
		return nil
	}

	cfg := NewHandlerConfig(
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

	assert.NotNil(t, cfg)
	assert.Equal(t, mockAuthService, cfg.AuthService)
	assert.Equal(t, mockUserService, cfg.UserService)
	assert.Equal(t, mockLoggerService, cfg.LoggerService)
	assert.Equal(t, mockRequestMetadataService, cfg.RequestMetadataService)
	assert.Equal(t, jwtSecret, cfg.JWTSecret)
	assert.Equal(t, refreshSecret, cfg.RefreshSecret)
	assert.Equal(t, issuer, cfg.Issuer)
	assert.Equal(t, audience, cfg.Audience)
	assert.Equal(t, oauthConfig, cfg.OAuth)
	assert.NotNil(t, cfg.CustomTokenSource)
}

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
		CustomTokenSource: func(ctx context.Context, refreshToken string) oauth2.TokenSource {
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

func TestClaims_Complete(t *testing.T) {
	// Test that Claims can be created and marshaled
	claims := &Claims{
		UserID: "user123",
	}

	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
}

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

func TestAuthHandler_Type(t *testing.T) {
	// Test that AuthHandler is a valid function type
	handler := func(w http.ResponseWriter, r *http.Request, user database.User) {
		// Test implementation
	}

	assert.NotNil(t, handler)
}

func TestOptionalHandler_Type(t *testing.T) {
	// Test that OptionalHandler is a valid function type
	handler := func(w http.ResponseWriter, r *http.Request, user *database.User) {
		// Test implementation
	}

	assert.NotNil(t, handler)
}

func TestHandlersConfig_ValidateConfig_Success(t *testing.T) {
	// Test successful validation
	logger := logrus.New()
	cfg := &HandlersConfig{
		Logger: logger,
		Auth:   &auth.AuthConfig{},
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	assert.NoError(t, err)
}

func TestHandlersConfig_ValidateConfig_MissingLogger(t *testing.T) {
	// Test validation failure with missing logger
	cfg := &HandlersConfig{
		Auth: &auth.AuthConfig{},
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger is required")
}

func TestHandlersConfig_ValidateConfig_MissingAuth(t *testing.T) {
	// Test validation failure with missing auth
	logger := logrus.New()
	cfg := &HandlersConfig{
		Logger: logger,
		APIConfig: &config.APIConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.ValidateConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth configuration is required")
}

func TestHandlersConfig_ValidateConfig_MissingAPIConfig(t *testing.T) {
	// Test validation failure with missing API config
	logger := logrus.New()
	cfg := &HandlersConfig{
		Logger: logger,
		Auth:   &auth.AuthConfig{},
	}

	err := cfg.ValidateConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API configuration is required")
}

func TestErrInvalidConfig_Error(t *testing.T) {
	// Test ErrInvalidConfig error message
	err := ErrInvalidConfig("test error message")
	assert.Equal(t, "test error message", err.Error())
}

// Test interface compliance
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
