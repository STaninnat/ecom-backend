// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// token_manager.go: JWT access/refresh token generation, storage, and validation.

// RedisRefreshTokenPrefix is the prefix used for refresh token keys in Redis.
const RedisRefreshTokenPrefix = "refresh_token:"

// GenerateAccessToken generates a signed JWT access token for the given user ID and expiration time.
func (cfg *Config) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	if cfg == nil {
		return "", errors.New("cfg is nil")
	}

	if err := ValidateConfig(cfg.JWTSecret, "JWTSecret"); err != nil {
		return "", err
	}

	timeNow := time.Now().UTC()
	if expiresAt.Before(timeNow) {
		return "", errors.New("expiresAt is in the past")
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Audience:  []string{cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("error signing JWT: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token for the given user ID using HMAC and a random UUID.
func (cfg *Config) GenerateRefreshToken(userID string) (string, error) {
	if cfg == nil {
		return "", errors.New("cfg is nil")
	}

	if err := ValidateConfig(cfg.RefreshSecret, "RefreshSecret"); err != nil {
		return "", err
	}

	rawUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error generating refresh token: %w", err)
	}

	message := fmt.Sprintf("%s:%s", userID, rawUUID.String())
	h := hmac.New(sha256.New, []byte(cfg.RefreshSecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s:%s:%s", userID, rawUUID.String(), signature), nil
}

// GenerateTokens generates both an access token and a refresh token for the given user ID.
func (cfg *Config) GenerateTokens(userID string, accessTokenExpiresAt time.Time) (string, string, error) {
	accessToken, err := cfg.GenerateAccessToken(userID, accessTokenExpiresAt)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := cfg.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// StoreRefreshTokenInRedis stores the refresh token and its metadata in Redis for the given user ID and provider.
func (cfg *Config) StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error {
	if cfg == nil {
		return errors.New("Config is nil")
	}
	if cfg.APIConfig == nil {
		return errors.New("APIConfig is nil")
	}
	if cfg.RedisClient == nil {
		return errors.New("RedisClient is nil")
	}

	if provider != "local" && provider != "google" {
		return fmt.Errorf("JSON Marshalling Error: unsupported provider %s", provider)
	}

	if ttl < 0 {
		return errors.New("invalid TTL")
	}

	if refreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}

	data := RefreshTokenData{
		Token:    refreshToken,
		Provider: provider,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Store refresh_token:<userID> -> token data (legacy)
	err = cfg.RedisClient.Set(r.Context(), RedisRefreshTokenPrefix+userID, jsonData, ttl).Err()
	if err != nil {
		return err
	}

	// Store refresh_token_lookup:<token> -> userID for O(1) lookup
	lookupKey := "refresh_token_lookup:" + refreshToken
	err = cfg.RedisClient.Set(r.Context(), lookupKey, userID, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

// ParseRefreshTokenData parses a JSON string into a RefreshTokenData struct and validates required fields.
func ParseRefreshTokenData(jsonData string) (RefreshTokenData, error) {
	var data RefreshTokenData

	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return RefreshTokenData{}, err
	}

	if data.Token == "" || data.Provider == "" {
		return RefreshTokenData{}, errors.New("token and provider fields are required")
	}

	return data, nil
}

// ValidateConfig checks that a secret is non-empty and of sufficient length.
func ValidateConfig(secret string, secretName string) error {
	if secret == "" {
		return fmt.Errorf("%s is empty", secretName)
	}

	if len(secret) < 32 {
		return fmt.Errorf("%s is too short", secretName)
	}

	return nil
}

// WARNING: GetUserIDFromRefreshToken uses Redis KEYS, which is slow for large datasets. Avoid in hot paths.

// GetUserIDByRefreshToken does an O(1) lookup for the user ID associated with a given refresh token in Redis.
func (cfg *Config) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	lookupKey := "refresh_token_lookup:" + refreshToken
	userID, err := cfg.RedisClient.Get(ctx, lookupKey).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}
