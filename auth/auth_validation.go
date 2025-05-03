package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func IsValidUserNameFormat(name string) bool {
	nameRegex := `^[a-zA-Z0-9]+([-._]?[a-zA-Z0-9]+)*$`

	re := regexp.MustCompile(nameRegex)

	return len(name) >= 3 && len(name) <= 30 && re.MatchString(name)
}

func IsValidEmailFormat(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+(?:\.[a-zA-Z0-9._%+-]+)*@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	if strings.Contains(email, "..") {
		return false
	}

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}

func (cfg *AuthConfig) ValidateAccessToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("invalid issuer: got '%s'", claims.Issuer)
	}

	if !slices.Contains(claims.Audience, cfg.Audience) {
		return nil, fmt.Errorf("invalid audience: got '%s'", claims.Audience)
	}

	timeNow := time.Now().UTC()
	if claims.ExpiresAt.Time.Before(timeNow) {
		return nil, fmt.Errorf("token expired")
	}

	if claims.NotBefore.Time.After(timeNow) {
		return nil, fmt.Errorf("token is not valid yet")
	}

	return claims, nil
}

func (cfg *AuthConfig) ValidateRefreshToken(refreshToken string) (uuid.UUID, error) {
	parts := strings.Split(refreshToken, ":")
	if len(parts) != 3 {
		userID, err := cfg.GetUserIDFromRefreshToken(refreshToken)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid refresh token format")
		}
		return userID, nil
	}

	userIDStr, rawUUID, signature := parts[0], parts[1], parts[2]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid userID in refresh token")
	}

	message := fmt.Sprintf("%s:%s", userIDStr, rawUUID)
	h := hmac.New(sha256.New, []byte(cfg.RefreshSecret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return uuid.Nil, fmt.Errorf("invalid refresh token signature")
	}

	return userID, nil
}

func (cfg *AuthConfig) ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (uuid.UUID, *RefreshTokenData, error) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return uuid.Nil, nil, err
	}
	refreshToken := cookie.Value

	userID, err := cfg.ValidateRefreshToken(refreshToken)
	if err != nil {
		return uuid.Nil, nil, err
	}

	storedTokenJSON, err := cfg.RedisClient.Get(r.Context(), "refresh_token:"+userID.String()).Result()
	if err != nil {
		return uuid.Nil, nil, err
	}

	storedData, err := ParseRefreshTokenData(storedTokenJSON)
	if err != nil {
		return uuid.Nil, nil, err
	}

	if storedData.Token != refreshToken {
		return uuid.Nil, nil, errors.New("invalid session")
	}

	return userID, &storedData, nil
}

func (cfg *AuthConfig) GetUserIDFromRefreshToken(refreshToken string) (uuid.UUID, error) {
	keys, err := cfg.RedisClient.Keys(context.Background(), "refresh_token:*").Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error fetching keys from Redis: %v", err)
	}

	for _, key := range keys {
		storedTokenJSON, err := cfg.RedisClient.Get(context.Background(), key).Result()
		if err != nil {
			continue
		}

		storedData, err := ParseRefreshTokenData(storedTokenJSON)
		if err != nil {
			continue
		}

		if storedData.Token == refreshToken && storedData.Provider == "google" {
			userIDStr := strings.TrimPrefix(key, "refresh_token:")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return uuid.Nil, fmt.Errorf("invalid user ID format in Redis key")
			}
			return userID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("refresh token not found in Redis")
}
