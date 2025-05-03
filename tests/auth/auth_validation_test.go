package auth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/go-redis/redismock/v9"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsValidUserNameFormat(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		// Valid cases
		{"Valid username with alphanumeric", "user123", true},
		{"Valid username with dash", "user-name", true},
		{"Valid username with underscore", "user_name", true},
		{"Valid username with period", "user.name", true},
		{"Valid username with mixed characters", "user.name123", true},
		{"Valid username at min length", "abc", true},                    // Minimum length boundary case
		{"Valid username at max length", "aVeryLongUserName12345", true}, // Maximum length boundary case

		// Invalid cases
		{"Invalid username too short", "ab", false},                               // Below minimum length
		{"Invalid username too long", "averylongusernamethatiswaytoolong", false}, // Exceeds maximum length
		{"Invalid username with spaces", "user name", false},                      // Spaces are not allowed
		{"Invalid username with special characters", "user!name", false},          // Special characters not allowed
		{"Invalid username with leading hyphen", "-username", false},              // Leading hyphen is not allowed
		{"Invalid username with leading period", ".username", false},              // Leading period is not allowed
		{"Invalid username with leading underscore", "_username", false},          // Leading underscore is not allowed
		{"Invalid username empty", "", false},                                     // Empty string should be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsValidUserNameFormat(tt.username)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for username %s", tt.expected, result, tt.username)
			}
		})
	}
}

func TestIsValidEmailFormat(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		// Valid cases
		{"Valid email with alphanumeric", "user123@example.com", true},
		{"Valid email with dot in domain", "user.name@example.com", true},
		{"Valid email with subdomain", "user@sub.example.com", true},
		{"Valid email with plus sign", "user+name@example.com", true}, // Plus sign is commonly used for filtering
		{"Valid email with hyphen in domain", "user-name@example-domain.com", true},
		{"Valid email with percentage sign", "user%name@example.com", true},

		// Invalid cases
		{"Invalid email with missing '@'", "userexample.com", false},             // Missing @ symbol
		{"Invalid email with missing domain", "user@.com", false},                // Missing domain name
		{"Invalid email with multiple '@'", "user@@example.com", false},          // Multiple @ symbols
		{"Invalid email with spaces", "user name@example.com", false},            // Spaces are not allowed
		{"Invalid email with special character", "user!name@example.com", false}, // Special characters are not allowed
		{"Invalid email with consecutive dots", "user..name@example.com", false}, // Consecutive dots are invalid
		{"Invalid email with domain only", "@example.com", false},                // Missing username
		{"Invalid email with invalid domain", "user@.com", false},                // Invalid domain format
		{"Invalid email with invalid characters", "user@#example.com", false},    // Invalid characters in domain
		{"Invalid email with empty string", "", false},                           // Empty string should be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsValidEmailFormat(tt.email)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for email %s", tt.expected, result, tt.email)
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	cfg := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			JWTSecret: "testsecrettestsecrettestsecretuwu",
			Issuer:    "test-issuer",
			Audience:  "test-audience",
		},
	}

	validUserID := uuid.New().String()
	validExpiry := time.Now().Add(1 * time.Hour)

	// Generate a valid token
	validToken, err := cfg.GenerateAccessToken(validUserID, validExpiry)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		tokenString string
		secret      string
		expectedErr string
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			secret:      cfg.JWTSecret,
			expectedErr: "",
		},
		{
			name:        "Invalid secret key",
			tokenString: validToken,
			secret:      "wrong-secret",
			expectedErr: "could not parse token", // Secret key mismatch
		},
		{
			name: "Expired token",
			tokenString: func() string {
				// Generate a token that has already expired
				expiredToken, _ := cfg.GenerateAccessToken(validUserID, time.Now().Add(-1*time.Hour))
				return expiredToken
			}(),
			secret:      cfg.JWTSecret,
			expectedErr: "could not parse token",
		},
		{
			name: "Token with invalid issuer",
			tokenString: func() string {
				// Generate a token with an incorrect issuer
				cfgWithInvalidIssuer := &auth.AuthConfig{
					APIConfig: &config.APIConfig{
						JWTSecret: cfg.JWTSecret,
						Issuer:    "wrong-issuer",
						Audience:  cfg.Audience,
					},
				}
				token, _ := cfgWithInvalidIssuer.GenerateAccessToken(validUserID, validExpiry)
				return token
			}(),
			secret:      cfg.JWTSecret,
			expectedErr: "invalid issuer",
		},
		{
			name: "Token with invalid audience",
			tokenString: func() string {
				// Generate a token with an incorrect audience
				cfgWithInvalidAudience := &auth.AuthConfig{
					APIConfig: &config.APIConfig{
						JWTSecret: cfg.JWTSecret,
						Issuer:    cfg.Issuer,
						Audience:  "wrong-audience",
					},
				}
				token, _ := cfgWithInvalidAudience.GenerateAccessToken(validUserID, validExpiry)
				return token
			}(),
			secret:      cfg.JWTSecret,
			expectedErr: "invalid audience",
		},
		{
			name: "Token not valid yet",
			tokenString: func() string {
				// Generate a token with a future NotBefore claim
				futureTime := time.Now().Add(1 * time.Hour)
				claims := auth.Claims{
					UserID: validUserID,
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.Issuer,
						Audience:  []string{cfg.Audience},
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						NotBefore: jwt.NewNumericDate(futureTime), // Token is not yet valid
						ExpiresAt: jwt.NewNumericDate(validExpiry),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			}(),
			secret:      cfg.JWTSecret,
			expectedErr: "token is not valid yet",
		},
		{
			name:        "Invalid token format",
			tokenString: "invalid.token.string",
			secret:      cfg.JWTSecret,
			expectedErr: "could not parse token", // Invalid structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := cfg.ValidateAccessToken(tt.tokenString, tt.secret)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, claims) // Expect valid claims for a valid token
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr) // Error should match expected message
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	client, mock := redismock.NewClientMock()

	cfg := auth.AuthConfig{
		APIConfig: &config.APIConfig{
			RefreshSecret: "testsecrettestsecrettestsecrettestsecretxdd",
			RedisClient:   client,
		},
	}

	validUserID := uuid.New().String()
	rawUUID := uuid.New()

	// Generate a valid refresh token for testing
	refreshToken, err := cfg.GenerateRefreshToken(validUserID)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		refreshToken  string
		redisResponse string
		expectedID    uuid.UUID
		expectedErr   string
	}{
		{
			name:         "Valid refresh token",
			refreshToken: refreshToken,
			redisResponse: fmt.Sprintf(`{
				"Token": "%s",
				"Provider": "google"
			}`, refreshToken),
			expectedID:  uuid.MustParse(validUserID),
			expectedErr: "",
		},
		{
			name:          "Invalid refresh token format",
			refreshToken:  "invalid-token-format",
			redisResponse: "",
			expectedID:    uuid.Nil,
			expectedErr:   "invalid refresh token format", // Token does not match expected format
		},
		{
			name:         "Invalid signature",
			refreshToken: fmt.Sprintf("%s:%s:%s", validUserID, rawUUID.String(), "invalid-signature"),
			redisResponse: fmt.Sprintf(`{
				"Token": "%s",
				"Provider": "google"
			}`, refreshToken),
			expectedID:  uuid.Nil,
			expectedErr: "invalid refresh token signature", // Signature mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Redis keys and values based on the test
			if tt.redisResponse != "" {
				mock.ExpectKeys("refresh_token:*").SetVal([]string{fmt.Sprintf("refresh_token:%s", validUserID)})
				mock.ExpectGet(fmt.Sprintf("refresh_token:%s", validUserID)).SetVal(tt.redisResponse)
			} else {
				mock.ExpectKeys("refresh_token:*").SetVal([]string{}) // No stored token
			}

			userID, err := cfg.ValidateRefreshToken(tt.refreshToken)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID) // Expect correct user ID for valid tokens
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr) // Error should match expected message
			}
		})
	}
}

func TestValidateCookieRefreshTokenData(t *testing.T) {
	client, mock := redismock.NewClientMock()

	cfg := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			RedisClient:   client,
			RefreshSecret: "testsecrettestsecrettestsecrettestsecretxdd",
		},
	}

	validUserID := uuid.New().String()
	validToken, err := cfg.GenerateRefreshToken(validUserID)
	assert.NoError(t, err)

	// Simulated valid stored token data in Redis
	validStoredData := auth.RefreshTokenData{
		Token:    validToken,
		Provider: "google",
	}
	validStoredJSON, _ := json.Marshal(validStoredData)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupRedis     func()
		expectedErrMsg string
	}{
		{
			name: "No refresh_token cookie",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},
			expectedErrMsg: "http: named cookie not present",
		},
		{
			name: "Invalid refresh_token format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid_token"})
				return req
			},
			expectedErrMsg: "invalid refresh token format",
		},
		{
			name: "Refresh token not found in Redis",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: validToken})
				return req
			},
			setupRedis: func() {
				mock.ExpectGet("refresh_token:" + validUserID).RedisNil() // Token does not exist
			},
			expectedErrMsg: "redis: nil",
		},
		{
			name: "Failed to unmarshal refresh token data",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: validToken})
				return req
			},
			setupRedis: func() {
				mock.ExpectGet("refresh_token:" + validUserID).SetVal("invalid_json") // Corrupt data
			},
			expectedErrMsg: "invalid character",
		},
		{
			name: "Refresh token mismatch",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: validToken})
				return req
			},
			setupRedis: func() {
				invalidData := auth.RefreshTokenData{
					Token:    "mismatched_token",
					Provider: "google",
				}
				invalidJSON, _ := json.Marshal(invalidData)
				mock.ExpectGet("refresh_token:" + validUserID).SetVal(string(invalidJSON))
			},
			expectedErrMsg: "invalid session",
		},
		{
			name: "Valid refresh token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: validToken})
				return req
			},
			setupRedis: func() {
				mock.ExpectGet("refresh_token:" + validUserID).SetVal(string(validStoredJSON)) // Valid session
			},
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := tt.setupRequest()
			if tt.setupRedis != nil {
				tt.setupRedis()
			}

			_, _, err := cfg.ValidateCookieRefreshTokenData(w, r)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserIDFromRefreshToken(t *testing.T) {
	client, mock := redismock.NewClientMock()

	cfg := &auth.AuthConfig{
		APIConfig: &config.APIConfig{
			RedisClient:   client,
			RefreshSecret: "testsecrettestsecrettestsecrettestsecretxdd",
		},
	}

	validToken := "valid_token"
	validProvider := "google"
	invalidProvider := "facebook"
	validUserID := uuid.New()

	validStoredData := auth.RefreshTokenData{
		Token:    validToken,
		Provider: validProvider,
	}
	validStoredJSON, _ := json.Marshal(validStoredData)

	tests := []struct {
		name         string
		setupRedis   func()
		refreshToken string
		expectedID   uuid.UUID
		expectedErr  string
	}{
		{
			name: "Error fetching keys from Redis",
			setupRedis: func() {
				// Simulate a Redis error when trying to fetch keys
				mock.ExpectKeys("refresh_token:*").SetErr(fmt.Errorf("redis error"))
			},
			refreshToken: validToken,
			expectedID:   uuid.Nil,
			expectedErr:  "error fetching keys from Redis",
		},
		{
			name: "Refresh token not found in Redis",
			setupRedis: func() {
				// Simulate an empty result from Redis, meaning the token doesn't exist
				mock.ExpectKeys("refresh_token:*").SetVal([]string{})
			},
			refreshToken: validToken,
			expectedID:   uuid.Nil,
			expectedErr:  "refresh token not found in Redis",
		},
		{
			name: "Provider mismatch in Redis",
			setupRedis: func() {
				// Store a refresh token but with a different provider (mismatch case)
				invalidData := auth.RefreshTokenData{
					Token:    validToken,
					Provider: invalidProvider, // Different provider
				}
				invalidJSON, _ := json.Marshal(invalidData)
				mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:" + validUserID.String()})
				mock.ExpectGet("refresh_token:" + validUserID.String()).SetVal(string(invalidJSON))
			},
			refreshToken: validToken,
			expectedID:   uuid.Nil,
			expectedErr:  "refresh token not found in Redis",
		},
		{
			name: "Invalid user ID format in Redis key",
			setupRedis: func() {
				// Store a token with an invalid UUID format in the Redis key
				mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:invalid_uuid"})
				mock.ExpectGet("refresh_token:invalid_uuid").SetVal(string(validStoredJSON))
			},
			refreshToken: validToken,
			expectedID:   uuid.Nil,
			expectedErr:  "invalid user ID format in Redis key",
		},
		{
			name: "Valid refresh token",
			setupRedis: func() {
				// Store a valid refresh token with correct provider in Redis
				mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:" + validUserID.String()})
				mock.ExpectGet("refresh_token:" + validUserID.String()).SetVal(string(validStoredJSON))
			},
			refreshToken: validToken,
			expectedID:   validUserID, // Expecting a valid user ID
			expectedErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupRedis()

			userID, err := cfg.GetUserIDFromRefreshToken(tt.refreshToken)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
