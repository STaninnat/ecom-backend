package auth

import (
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthConfig struct {
	*config.APIConfig
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}

type AuthService interface {
	GenerateAccessToken(userID uuid.UUID, secret string, expiresAt time.Time) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string, secret string) (*Claims, error)
	ValidateRefreshToken(refreshToken string) (uuid.UUID, error)
	StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error
}
