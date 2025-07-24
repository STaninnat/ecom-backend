// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/STaninnat/ecom-backend/internal/config"
)

// auth_embedded.go: Embedded authentication configuration and helpers.

// Config wraps the APIConfig for authentication-related configuration.
type Config struct {
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
