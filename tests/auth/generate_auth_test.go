package auth

import (
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hashedPassword, err := auth.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	password := ""
	hashedPassword, err := auth.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestHashPassword_Error(t *testing.T) {
	password := string(make([]byte, 10000))
	hashedPassword, err := auth.HashPassword(password)

	assert.Error(t, err)
	assert.Empty(t, hashedPassword)
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	password := "samePassword"

	hashedPassword1, err1 := auth.HashPassword(password)
	hashedPassword2, err2 := auth.HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.NotEqual(t, hashedPassword1, hashedPassword2)
}

func TestGenerateAccessToken(t *testing.T) {
	authConfig := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			Issuer:   "testIssuer",
			Audience: "testAudience",
		},
	}

	userID := uuid.New()
	secret := "secret"
	expiresAt := time.Now().Add(time.Hour)

	token, err := authConfig.GenerateAccessToken(userID, secret, expiresAt)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(token, &auth.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	assert.NoError(t, err)
	claims, ok := parsedToken.Claims.(*auth.Claims)
	assert.True(t, ok)

	assert.Equal(t, userID, claims.UserID)
}

func TestGenerateAccessTokenWithNilConfig(t *testing.T) {
	authConfig := (*auth.AuthConfig)(nil)

	userID := uuid.New()
	secret := "secret"
	expiresAt := time.Now().Add(time.Hour)

	_, err := authConfig.GenerateAccessToken(userID, secret, expiresAt)

	assert.Error(t, err)
	assert.Equal(t, "cfg is nil", err.Error())
}

func TestGenerateRefreshToken(t *testing.T) {
	authConfig := &auth.AuthConfig{}
	refreshToken, err := authConfig.GenerateRefreshToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
}
