package auth

import (
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// NOTE: This file only contains type definitions and imports. No performance improvements needed.

// AuthConfig wraps the APIConfig for authentication-related configuration.
type AuthConfig struct {
	*config.APIConfig
}

// Claims represents the JWT claims used for authentication, including the user ID and standard registered claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// RefreshTokenData holds information about a refresh token and its associated provider.
type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}
