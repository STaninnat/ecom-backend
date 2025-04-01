package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (cfg *AuthConfig) GenerateAccessToken(userID uuid.UUID, expiresAt time.Time) (string, error) {
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

func (cfg *AuthConfig) GenerateRefreshToken(userID uuid.UUID) (string, error) {
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

	message := fmt.Sprintf("%s:%s", userID.String(), rawUUID.String())
	h := hmac.New(sha256.New, []byte(cfg.RefreshSecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s:%s:%s", userID.String(), rawUUID.String(), signature), nil
}

func (cfg *AuthConfig) GenerateTokens(userID uuid.UUID, accessTokenExpiresAt time.Time) (string, string, error) {
	accessToken, err := cfg.GenerateAccessToken(userID, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error generating access token:", err)
		return "", "", err
	}

	newRefreshToken, err := cfg.GenerateRefreshToken(userID)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (cfg *AuthConfig) StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error {
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
		log.Println("Failed to marshal refresh token data:", err)
		return err
	}

	err = cfg.RedisClient.Set(r.Context(), "refresh_token:"+userID, jsonData, ttl).Err()
	if err != nil {
		log.Println("Failed to save refresh token to Redis:", err)
		return err
	}

	return nil
}

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

func ValidateConfig(secret string, secretName string) error {
	if secret == "" {
		return fmt.Errorf("%s is empty", secretName)
	}

	if len(secret) < 32 {
		return fmt.Errorf("%s is too short", secretName)
	}

	return nil
}
