package handlers

import (
	"log"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/sirupsen/logrus"
)

type HandlersConfig struct {
	*config.APIConfig
	Auth       *auth.AuthConfig
	OAuth      *config.OAuthConfig
	AuthHelper auth.AuthHelper
	Logger     *logrus.Logger
}

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
		AuthHelper: &auth.RealHelper{
			AuthConfig: authCfg,
		},
		Logger: logger,
	}
}
