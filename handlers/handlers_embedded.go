// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"context"
	"log"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// handlers_embedded.go: Defines the main handler configuration struct and helper functions to initialize and validate handler dependencies.

// Config represents the main configuration for handlers, including API, auth, OAuth, logger, and cache service.
type Config struct {
	*config.APIConfig
	Auth              *auth.Config
	OAuth             *config.OAuthConfig
	Logger            *logrus.Logger
	CustomTokenSource func(ctx context.Context, refreshToken string) oauth2.TokenSource
	CacheService      *utils.CacheService
}

// HandlerResponse represents a standard handler response with a message.
type HandlerResponse struct {
	Message string `json:"message"`
}

// TokenTTL constants define the default TTL for access and refresh tokens.
const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// SetupHandlersConfig creates and configures a new HandlersConfig instance with all dependencies.
func SetupHandlersConfig(logger *logrus.Logger) *Config {
	apicfg := config.LoadConfig()

	// Connect to database
	apicfg.ConnectDB()

	// Load OAuth configuration with error handling
	oauthConfig, err := config.NewOAuthConfig(apicfg.CredsPath)
	if err != nil {
		log.Fatal("Failed to load oauth config: ", err)
	}

	// Create auth configuration
	authCfg := &auth.Config{
		APIConfig: apicfg,
	}

	// Create cache service
	cacheService := utils.NewCacheService(apicfg.RedisClient)

	return &Config{
		APIConfig:    apicfg,
		Auth:         authCfg,
		OAuth:        oauthConfig,
		Logger:       logger,
		CacheService: cacheService,
	}
}

// NewHandlerConfig creates a new HandlerConfig with interfaces for better testability.
func NewHandlerConfig(
	authService AuthService,
	userService UserService,
	loggerService LoggerService,
	requestMetadataService RequestMetadataService,
	jwtSecret, refreshSecret, issuer, audience string,
	oauth *OAuthConfig,
	customTokenSource func(ctx context.Context, refreshToken string) oauth2.TokenSource,
) *HandlerConfig {
	return &HandlerConfig{
		AuthService:            authService,
		UserService:            userService,
		LoggerService:          loggerService,
		RequestMetadataService: requestMetadataService,
		JWTSecret:              jwtSecret,
		RefreshSecret:          refreshSecret,
		Issuer:                 issuer,
		Audience:               audience,
		OAuth:                  oauth,
		CustomTokenSource:      customTokenSource,
	}
}

// ValidateConfig validates the handler configuration and returns an error if invalid.
func (cfg *Config) ValidateConfig() error {
	if cfg.Logger == nil {
		return ErrInvalidConfig("logger is required")
	}
	if cfg.Auth == nil {
		return ErrInvalidConfig("auth configuration is required")
	}
	if cfg.APIConfig == nil {
		return ErrInvalidConfig("API configuration is required")
	}
	return nil
}

// ErrInvalidConfig represents an invalid configuration error.
type ErrInvalidConfig string

// Error implements the error interface for ErrInvalidConfig.
func (e ErrInvalidConfig) Error() string {
	return string(e)
}
