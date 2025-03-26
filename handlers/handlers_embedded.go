package handlers

import (
	"log"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
)

type HandlersConfig struct {
	*config.APIConfig
	Auth  *auth.AuthConfig
	OAuth *config.OAuthConfig
}

func SetupHandlersConfig() *HandlersConfig {
	apicfg := config.LoadConfig()
	apicfg.ConnectDB()

	oauthConfig, err := config.NewOAuthConfig()
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
	}
}
