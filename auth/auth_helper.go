package auth

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RealHelper struct {
	*AuthConfig
}

type AuthHelper interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) (bool, error)
	GenerateTokens(userID string, expiresAt time.Time) (string, string, error)
	StoreRefreshTokenInRedis(r *http.Request, userID, token, provider string, ttl time.Duration) error
	ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (uuid.UUID, *RefreshTokenData, error)
}

func (r *RealHelper) HashPassword(password string) (string, error) {
	return HashPassword(password)
}

func (r *RealHelper) CheckPasswordHash(password, hash string) (bool, error) {
	return CheckPasswordHash(password, hash)
}

func (r *RealHelper) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return r.AuthConfig.GenerateTokens(userID, expiresAt)
}

func (r *RealHelper) StoreRefreshTokenInRedis(rq *http.Request, userID, token, provider string, ttl time.Duration) error {
	return r.AuthConfig.StoreRefreshTokenInRedis(rq, userID, token, provider, ttl)
}

func (r *RealHelper) ValidateCookieRefreshTokenData(w http.ResponseWriter, rq *http.Request) (uuid.UUID, *RefreshTokenData, error) {
	return r.AuthConfig.ValidateCookieRefreshTokenData(w, rq)
}
