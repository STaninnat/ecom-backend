package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestIsValidUserNameFormat(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"valid name", "user_123", true},
		{"too short name", "us", false},
		{"too long name", "aVeryLongUserNameThatExceeds30Chars", false},
		{"valid name with dot", "user.name", true},
		{"invalid name with space", "user name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsValidUserNameFormat(tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidateEmailFormat(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"user@example.com", true},
		{"invalid-email", false},
		{"user@domain.co", true},
		{"user@domain", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := auth.IsValidEmailFormat(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "securePassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"correct password", password, true},
		{"incorrect password", "wrongPassword", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.CheckPasswordHash(tt.password, string(hash))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	secret := "secret"
	invalidToken := "invalidToken"

	authConfig := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			Issuer:   "testIssuer",
			Audience: "testAudience",
		},
	}

	validClaims := auth.Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    authConfig.Issuer,
			Audience:  []string{authConfig.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	expiredClaims := auth.Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    authConfig.Issuer,
			Audience:  []string{authConfig.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	invalidIssuerClaims := auth.Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "invalidIssuer",
			Audience:  []string{authConfig.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	invalidIssuerToken := jwt.NewWithClaims(jwt.SigningMethodHS256, invalidIssuerClaims)
	invalidIssuerTokenString, err := invalidIssuerToken.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token with invalid issuer: %v", err)
	}

	invalidAudienceClaims := auth.Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    authConfig.Issuer,
			Audience:  []string{"invalidAudience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	invalidAudienceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, invalidAudienceClaims)
	invalidAudienceTokenString, err := invalidAudienceToken.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token with invalid audience: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		secret      string
		expectedErr string
	}{
		{
			name:        "Valid token",
			tokenString: tokenString,
			secret:      secret,
			expectedErr: "",
		},
		{
			name:        "Invalid token",
			tokenString: invalidToken,
			secret:      secret,
			expectedErr: "could not parse token",
		},
		{
			name:        "Expired token",
			tokenString: expiredTokenString,
			secret:      secret,
			expectedErr: "token has invalid claims: token is expired",
		},
		{
			name:        "Invalid issuer",
			tokenString: invalidIssuerTokenString,
			secret:      secret,
			expectedErr: "invalid issuer",
		},
		{
			name:        "Invalid audience",
			tokenString: invalidAudienceTokenString,
			secret:      secret,
			expectedErr: "invalid audience",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authConfig.ValidateAccessToken(tt.tokenString, tt.secret)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	cfg := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			RefreshSecret: "test-secret",
		},
	}

	userID := uuid.New()
	validToken, err := cfg.GenerateRefreshToken(userID)
	assert.NoError(t, err, "Should not error when generating refresh token")

	t.Run("Valid Token", func(t *testing.T) {
		parsedUserID, err := cfg.ValidateRefreshToken(validToken)
		assert.NoError(t, err, "Should not error when validating valid refresh token")
		assert.Equal(t, userID, parsedUserID, "UserID should match")
	})

	t.Run("Invalid Format", func(t *testing.T) {
		invalidToken := "invalid:token"
		parsedUserID, err := cfg.ValidateRefreshToken(invalidToken)
		assert.Error(t, err, "Should error on invalid token format")
		assert.Equal(t, uuid.Nil, parsedUserID, "UserID should be nil")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		invalidToken := validToken[:len(validToken)-1] + "x"
		parsedUserID, err := cfg.ValidateRefreshToken(invalidToken)
		assert.Error(t, err, "Should error on invalid token signature")
		assert.Equal(t, uuid.Nil, parsedUserID, "UserID should be nil")
	})
}

type TestDocodeStruct struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func TestDecodeAndValidate(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		expectedResponse int
		expectValid      bool
	}{
		{"Valid JSON", `{"field1": "value1", "field2": "value2"}`, http.StatusOK, true},
		{"Invalid JSON", `{"field1": "`, http.StatusBadRequest, false},
		{"Missing field", `{"field1": ""}`, http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			params, valid := auth.DecodeAndValidate[TestDocodeStruct](w, req)

			if valid != tt.expectValid {
				t.Errorf("expected valid to be %v, got %v", tt.expectValid, valid)
			}

			if status := w.Result().StatusCode; status != tt.expectedResponse {
				t.Errorf("expected status %v, got %v", tt.expectedResponse, status)
			}

			if tt.expectValid && params == nil {
				t.Errorf("expected non-nil params, got nil")
			}
		})
	}
}
