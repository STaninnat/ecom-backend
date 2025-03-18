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

func IsValidEmailFormat(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
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

	timeNow := time.Now().Local()
	if claims.ExpiresAt.Time.Before(timeNow) {
		return nil, fmt.Errorf("token expired")
	}

	if claims.NotBefore.Time.After(timeNow) {
		return nil, fmt.Errorf("token not valid yet")
	}

	return claims, nil
}
