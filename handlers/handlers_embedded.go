package handlers

import (
	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
)

type HandlersConfig struct {
	*config.APIConfig
	Auth *auth.AuthConfig
}
