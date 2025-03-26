package auth

import (
	"time"

	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/google/uuid"
)

type AuthConfig struct {
	*config.APIConfig
}

type AuthService interface {
	GenerateAccessToken(userID uuid.UUID, secret string, expiresAt time.Time) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string, secret string) (*Claims, error)
	ValidateRefreshToken(refreshToken string) (uuid.UUID, error)
}
