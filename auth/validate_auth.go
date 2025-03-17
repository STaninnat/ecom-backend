package auth

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func IsValidUserNameFormat(name string) bool {
	nameRegex := `^[a-zA-Z0-9]+([-._]?[a-zA-Z0-9]+)*$`

	re := regexp.MustCompile(nameRegex)

	return len(name) >= 3 && len(name) <= 30 && re.MatchString(name)
}

func IsValidateEmailFormat(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}

func CheckPasswordHash(password, hash string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}

	return true
}

func (cfg *AuthConfig) ValidateAccessToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected 'my-api-service', got '%s'", claims.Issuer)
	}

	if !contains(claims.Audience, cfg.Audience) {
		return nil, fmt.Errorf("invalid audience: expected 'my-frontend-app', got '%s'", claims.Audience)
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	if claims.NotBefore.After(time.Now()) {
		return nil, fmt.Errorf("token not valid yet")
	}

	return claims, nil
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
