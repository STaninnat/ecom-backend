package mocks

import (
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/google/uuid"
)

type MockAuthHelper struct {
	HashPasswordFn                   func(string) (string, error)
	CheckPasswordHashFn              func(string, string) error
	GenerateTokensFn                 func(string, time.Time) (string, string, error)
	StoreRefreshTokenInRedisFn       func(*http.Request, string, string, string, time.Duration) error
	ValidateCookieRefreshTokenDataFn func(http.ResponseWriter, *http.Request) (uuid.UUID, *auth.RefreshTokenData, error)
}

func (m *MockAuthHelper) HashPassword(password string) (string, error) {
	if m.HashPasswordFn != nil {
		return m.HashPasswordFn(password)
	}
	return "hashed-password", nil
}

func (m *MockAuthHelper) CheckPasswordHash(password, hash string) error {
	if m.CheckPasswordHashFn != nil {
		return m.CheckPasswordHashFn(password, hash)
	}
	return nil
}

func (m *MockAuthHelper) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	if m.GenerateTokensFn != nil {
		return m.GenerateTokensFn(userID, expiresAt)
	}
	return "access-token", "refresh-token", nil
}

func (m *MockAuthHelper) StoreRefreshTokenInRedis(r *http.Request, userID, token, provider string, ttl time.Duration) error {
	if m.StoreRefreshTokenInRedisFn != nil {
		return m.StoreRefreshTokenInRedisFn(r, userID, token, provider, ttl)
	}
	return nil
}

func (m *MockAuthHelper) ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
	if m.ValidateCookieRefreshTokenDataFn != nil {
		return m.ValidateCookieRefreshTokenDataFn(w, r)
	}
	return uuid.Nil, &auth.RefreshTokenData{}, nil
}
