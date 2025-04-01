package auth_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/go-redis/redismock/v9"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var NewRandomUUID = uuid.NewRandom

func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *auth.AuthConfig
		userID      uuid.UUID
		expiresAt   time.Time
		expectedErr error
	}{
		{
			name: "valid config and parameters",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:    "testIssuer",
					Audience:  "testAudience",
					JWTSecret: "validsecretvalidsecretvalidsecretuwu",
				},
			},
			userID:      uuid.New(),
			expiresAt:   time.Now().Add(time.Hour),
			expectedErr: nil,
		},
		{
			name:        "nil config",
			cfg:         nil,
			userID:      uuid.New(),
			expiresAt:   time.Now().Add(time.Hour),
			expectedErr: errors.New("cfg is nil"),
		},
		{
			name: "error signing token (invalid JWTSecret)",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:    "testIssuer",
					Audience:  "testAudience",
					JWTSecret: "",
				},
			},
			userID:      uuid.New(),
			expiresAt:   time.Now().Add(time.Hour),
			expectedErr: errors.New("JWTSecret is empty"),
		},
		{
			name: "error generating token (expiresAt in the past)",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:    "testIssuer",
					Audience:  "testAudience",
					JWTSecret: "validsecretvalidsecretvalidsecretuwu",
				},
			},
			userID:      uuid.New(),
			expiresAt:   time.Now().Add(-time.Hour),
			expectedErr: errors.New("expiresAt is in the past"),
		},
		{
			name: "error signing token (invalid JWTSecret format)",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:    "testIssuer",
					Audience:  "testAudience",
					JWTSecret: "123",
				},
			},
			userID:      uuid.New(),
			expiresAt:   time.Now().Add(time.Hour),
			expectedErr: errors.New("JWTSecret is too short"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.cfg.GenerateAccessToken(tt.userID, tt.expiresAt)

			// Check if an error is expected
			if tt.expectedErr != nil {
				assert.Error(t, err)                                    // Ensure there is an error
				assert.Contains(t, err.Error(), tt.expectedErr.Error()) // Verify the error message contains expected text
				return                                                  // Skip further checks since we expect an error
			}

			// If no error is expected, validate the generated token
			assert.NoError(t, err)    // Ensure no error occurred
			assert.NotEmpty(t, token) // Token should not be empty

			parsedToken, parseErr := jwt.ParseWithClaims(token, &auth.Claims{}, func(token *jwt.Token) (any, error) {
				return []byte(tt.cfg.JWTSecret), nil // Use the JWTSecret to verify the token signature
			})
			assert.NoError(t, parseErr) // Ensure parsing does not result in an error

			// Extract claims and verify the userID is correct
			claims, ok := parsedToken.Claims.(*auth.Claims)
			assert.True(t, ok)                        // Ensure claims are of the correct type
			assert.Equal(t, tt.userID, claims.UserID) // Ensure the userID in the token matches the one provided
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *auth.AuthConfig
		userID      uuid.UUID
		expectedErr error
		expectedTok string
	}{
		{
			name: "valid config and parameters",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RefreshSecret: "validsecretvalidsecretvalidsecretxdd",
				},
			},
			userID:      uuid.New(),
			expectedErr: nil,
		},
		{
			name:        "nil config",
			cfg:         nil,
			userID:      uuid.New(),
			expectedErr: errors.New("cfg is nil"),
		},
		{
			name: "empty RefreshSecret",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RefreshSecret: "",
				},
			},
			userID:      uuid.New(),
			expectedErr: errors.New("RefreshSecret is empty"),
		},
		{
			name: "invalid RefreshSecret length",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RefreshSecret: "short",
				},
			},
			userID:      uuid.New(),
			expectedErr: errors.New("RefreshSecret is too short"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			token, err := tt.cfg.GenerateRefreshToken(tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			parts := strings.Split(token, ":")
			assert.Len(t, parts, 3, "Token format should be 'userID:uuid:signature'")

			_, err = uuid.Parse(parts[1])
			assert.NoError(t, err, "Invalid UUID format")

			assert.True(t, len(parts[2]) > 0, "Signature part should not be empty")
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	tests := []struct {
		name                    string
		cfg                     *auth.AuthConfig
		userID                  uuid.UUID
		accessTokenExpiresAt    time.Time
		mockAccessTokenFailure  bool
		mockRefreshTokenFailure bool
		expectedErr             error
	}{
		{
			name: "valid tokens generation",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:        "testIssuer",
					Audience:      "testAudience",
					JWTSecret:     "validsecretvalidsecretvalidsecretuwu",
					RefreshSecret: "validsecretvalidsecretvalidsecretxdd",
				},
			},
			userID:                  uuid.New(),
			accessTokenExpiresAt:    time.Now().Add(time.Hour),
			mockAccessTokenFailure:  false,
			mockRefreshTokenFailure: false,
			expectedErr:             nil,
		},
		{
			name: "error generating access token",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:        "testIssuer",
					Audience:      "testAudience",
					JWTSecret:     "",
					RefreshSecret: "validsecretvalidsecretvalidsecretxdd",
				},
			},
			userID:                  uuid.New(),
			accessTokenExpiresAt:    time.Now().Add(time.Hour),
			mockAccessTokenFailure:  true,
			mockRefreshTokenFailure: false,
			expectedErr:             errors.New("JWTSecret is empty"),
		},
		{
			name: "error generating refresh token",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					Issuer:        "testIssuer",
					Audience:      "testAudience",
					JWTSecret:     "validsecretvalidsecretvalidsecretuwu",
					RefreshSecret: "",
				},
			},
			userID:                  uuid.New(),
			accessTokenExpiresAt:    time.Now().Add(time.Hour),
			mockAccessTokenFailure:  false,
			mockRefreshTokenFailure: true,
			expectedErr:             errors.New("RefreshSecret is empty"),
		},
	}

	// Store the original UUID generator function to restore later
	originalUUIDGenerator := NewRandomUUID
	defer func() { NewRandomUUID = originalUUIDGenerator }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockRefreshTokenFailure {
				NewRandomUUID = func() (uuid.UUID, error) {
					return uuid.UUID{}, errors.New("UUID generation failed")
				}
			} else {
				NewRandomUUID = originalUUIDGenerator
			}

			accessToken, refreshToken, err := tt.cfg.GenerateTokens(tt.userID, tt.accessTokenExpiresAt)

			if tt.expectedErr != nil {
				assert.Error(t, err)                                    // Ensure an error occurred
				assert.Contains(t, err.Error(), tt.expectedErr.Error()) // Verify the error message
				assert.Empty(t, accessToken)                            // Access token should be empty
				assert.Empty(t, refreshToken)                           // Refresh token should be empty
				return                                                  // Skip further checks as we expect an error
			}

			// If no error is expected, validate the generated tokens
			assert.NoError(t, err)           // Ensure no error occurred
			assert.NotEmpty(t, accessToken)  // Ensure the access token is not empty
			assert.NotEmpty(t, refreshToken) // Ensure the refresh token is not empty

			parts := strings.Split(refreshToken, ":")
			assert.Len(t, parts, 3, "Refresh token format should be 'userID:uuid:signature'")
		})
	}
}

func TestStoreRefreshTokenInRedis(t *testing.T) {
	client, mock := redismock.NewClientMock()
	tests := []struct {
		name         string
		cfg          *auth.AuthConfig
		userID       string
		refreshToken string
		provider     string
		ttl          time.Duration
		expectedErr  error
	}{
		{
			name: "Success",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "google",
			ttl:          10 * time.Minute,
			expectedErr:  nil,
		},
		{
			name: "JSON Marshalling Error",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "invalid_provider",
			ttl:          10 * time.Minute,
			expectedErr:  errors.New("JSON Marshalling Error:"),
		},
		{
			name: "Redis Set Error",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "google",
			ttl:          10 * time.Minute,
			expectedErr:  errors.New("Redis error"),
		},
		{
			name: "Redis Client is nil",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: nil,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "google",
			ttl:          10 * time.Minute,
			expectedErr:  errors.New("RedisClient is nil"),
		},
		{
			name: "Invalid TTL",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "google",
			ttl:          -1 * time.Minute,
			expectedErr:  errors.New("invalid TTL"),
		},
		{
			name: "Empty Refresh Token",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "",
			provider:     "google",
			ttl:          10 * time.Minute,
			expectedErr:  errors.New("refresh token cannot be empty"),
		},
		{
			name: "Successful Redis Set with TTL Zero",
			cfg: &auth.AuthConfig{
				APIConfig: &config.APIConfig{
					RedisClient: client,
				},
			},
			userID:       "user123",
			refreshToken: "validsecretvalidsecretvalidsecretxdd",
			provider:     "google",
			ttl:          0,
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare mock Redis client for each test case
			if tt.cfg.APIConfig.RedisClient != nil {
				tt.cfg.APIConfig.RedisClient = client
			}

			// Simulate the data
			data := auth.RefreshTokenData{
				Token:    tt.refreshToken,
				Provider: tt.provider,
			}

			// Check for marshaling errors
			jsonData, err := json.Marshal(data)
			if err != nil && tt.expectedErr == nil {
				// If marshalling fails, ensure no Redis call is made
				t.Fatal("Unexpected marshalling error")
			}

			if tt.expectedErr == nil {
				// Success case: Mock Redis expectation
				mock.ExpectSet("refresh_token:"+tt.userID, jsonData, tt.ttl).SetVal("OK")
			} else if tt.expectedErr.Error() == "Redis error" {
				// Redis error case: Mock Redis error
				mock.ExpectSet("refresh_token:"+tt.userID, jsonData, tt.ttl).SetErr(errors.New("Redis error"))
			}

			err = tt.cfg.StoreRefreshTokenInRedis(&http.Request{}, tt.userID, tt.refreshToken, tt.provider, tt.ttl)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestParseRefreshTokenData(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		expected  auth.RefreshTokenData
		expectErr bool
	}{
		{
			name:     "Valid JSON",
			jsonData: `{"token":"abc123","provider":"google"}`,
			expected: auth.RefreshTokenData{Token: "abc123", Provider: "google"},
		},
		{
			name:      "Missing token",
			jsonData:  `{"provider":"google"}`,
			expectErr: true,
		},
		{
			name:      "Missing provider",
			jsonData:  `{"token":"abc123"}`,
			expectErr: true,
		},
		{
			name:      "Empty JSON object",
			jsonData:  `{}`,
			expectErr: true,
		},
		{
			name:      "Invalid JSON format",
			jsonData:  `{"token":"abc123", "provider":"google"`, // Invalid JSON format (missing closing brace)
			expectErr: true,
		},
		{
			name:      "Empty string",
			jsonData:  ``,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := auth.ParseRefreshTokenData(tt.jsonData)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got error: %v", tt.expectErr, err)
			}

			if !tt.expectErr && result != tt.expected {
				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		secretName  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid secret",
			secret:      "thisisaverylongsecretkeythatexceedslimit",
			secretName:  "API_KEY",
			expectError: false,
		},
		{
			name:        "Empty secret",
			secret:      "",
			secretName:  "API_KEY",
			expectError: true,
			errorMsg:    "API_KEY is empty",
		},
		{
			name:        "Short secret",
			secret:      "shortsecret",
			secretName:  "JWT_SECRET",
			expectError: true,
			errorMsg:    "JWT_SECRET is too short",
		},
		{
			name:        "Exactly 32 characters",
			secret:      "12345678901234567890123456789012",
			secretName:  "SESSION_SECRET",
			expectError: false,
		},
		{
			name:        "Just below 32 characters",
			secret:      "1234567890123456789012345678901",
			secretName:  "DB_PASSWORD",
			expectError: true,
			errorMsg:    "DB_PASSWORD is too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateConfig(tt.secret, tt.secretName)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			// If error occurred, check if the error message matches the expected message
			if err != nil && err.Error() != tt.errorMsg {
				t.Errorf("expected error message: %q, got: %q", tt.errorMsg, err.Error())
			}
		})
	}
}
