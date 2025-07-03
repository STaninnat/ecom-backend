package handlers

import (
	"context"
	"log"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type HandlersConfig struct {
	*config.APIConfig
	Auth              *auth.AuthConfig
	OAuth             *config.OAuthConfig
	Logger            *logrus.Logger
	CustomTokenSource func(ctx context.Context, refreshToken string) oauth2.TokenSource
}

type HandlerResponse struct {
	Message string `json:"message"`
}

const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func SetupHandlersConfig(logger *logrus.Logger) *HandlersConfig {
	apicfg := config.LoadConfig()
	apicfg.ConnectDB()

	oauthConfig, err := config.NewOAuthConfig(apicfg.CredsPath)
	if err != nil {
		log.Fatal("Failed to load oauth config: ", err)
	}

	authCfg := &auth.AuthConfig{
		APIConfig: apicfg,
	}

	return &HandlersConfig{
		APIConfig: apicfg,
		Auth:      authCfg,
		OAuth:     oauthConfig,
		Logger:    logger,
	}
}
