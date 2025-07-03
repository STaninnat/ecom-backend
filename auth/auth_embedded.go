package auth

import (
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// NOTE: This file only contains type definitions and imports. No performance improvements needed.

type AuthConfig struct {
	*config.APIConfig
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}
