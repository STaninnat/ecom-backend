package config

import "github.com/STaninnat/ecom-backend/internal/database"

type APIConfig struct {
	DB        *database.Queries
	JWTSecret string
	Issuer    string
	Audience  string
}
