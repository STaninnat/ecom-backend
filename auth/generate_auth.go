package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Couldn't hash password")
		return "", err
	}

	return string(bytes), nil
}

func (cfg *AuthConfig) GenerateAccessToken(userID uuid.UUID, secret string, expiresAt time.Time) (string, error) {
	if cfg == nil {
		return "", errors.New("cfg is nil")
	}

	timeNow := time.Now().UTC()

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
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error signing JWT: %w", err)
	}

	return tokenString, nil
}

func (cfg *AuthConfig) GenerateRefreshToken(userID uuid.UUID) (string, error) {
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
