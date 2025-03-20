package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupAPICfg(mockRedis *redis.Client, mockDB *database.Queries) *handlers.HandlersConfig {
	return &handlers.HandlersConfig{
		APIConfig: &config.APIConfig{
			DB:          mockDB,
			RedisClient: mockRedis,
			JWTSecret:   "test-secret",
		},
		Auth: &auth.AuthConfig{},
	}
}

func runSignUpTest(t *testing.T, apicfg *handlers.HandlersConfig, reqBody map[string]string, expectedStatus int, expectedMessage string) {
	reqJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apicfg.HandlerSignUp(w, req)

	require.Equal(t, expectedStatus, w.Code)

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	actualMessage, hasMessage := resp["message"]
	actualError, hasError := resp["error"]

	if hasMessage {
		require.Equal(t, expectedMessage, actualMessage)
	} else if hasError {
		require.Equal(t, expectedMessage, actualError)
	} else {
		t.Fatalf("Neither 'message' nor 'error' found in response")
	}
}

func setupMockFunctions(someCondition, hashPasswordErr, generateAccessTokenErr, generateRefreshTokenErr bool) *gomonkey.Patches {
	patches := gomonkey.NewPatches().
		ApplyFunc((*database.Queries).CheckUserExistsByName, func(_ *database.Queries, _ context.Context, name string) (bool, error) {
			if name == "existing_user" {
				return true, nil
			}
			return false, nil
		}).
		ApplyFunc((*database.Queries).CheckUserExistsByEmail, func(_ *database.Queries, _ context.Context, email string) (bool, error) {
			if email == "existing@example.com" {
				return true, nil
			}
			return false, nil
		}).
		ApplyFunc((*database.Queries).CreateUser, func(_ *database.Queries, _ context.Context, _ database.CreateUserParams) error {
			if someCondition {
				return errors.New("database error")
			}
			return nil
		}).
		ApplyFunc((*auth.AuthConfig).GenerateAccessToken, func(_ *auth.AuthConfig, _ uuid.UUID, _ string, _ time.Time) (string, error) {
			if generateAccessTokenErr {
				return "", errors.New("failed to generate access token")
			}
			return "mockAccessToken", nil
		}).
		ApplyFunc((*auth.AuthConfig).GenerateRefreshToken, func(_ *auth.AuthConfig) (string, error) {
			if generateRefreshTokenErr {
				return "", errors.New("failed to generate refresh token")
			}
			return "mockRefreshToken", nil
		}).
		ApplyFunc(auth.HashPassword, func(password string) (string, error) {
			if hashPasswordErr {
				return "", errors.New("password hashing error")
			}
			return "hashedPassword", nil
		})

	return patches
}

func TestHandlerSignUp(t *testing.T) {
	mockRedis, mockRedisClient := redismock.NewClientMock()
	mockDB := &database.Queries{}
	apicfg := setupAPICfg(mockRedis, mockDB)

	tests := []struct {
		name                    string
		reqBody                 map[string]string
		someCondition           bool
		hashPasswordErr         bool
		generateAccessTokenErr  bool
		generateRefreshTokenErr bool
		redisError              bool
		expectedStatus          int
		expectedMessage         string
	}{
		{
			name:            "Invalid JSON format",
			reqBody:         map[string]string{"name": "new_user"},
			someCondition:   false,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid request format",
		},
		{
			name:            "Successful Signup",
			reqBody:         map[string]string{"name": "new_user", "email": "new@example.com", "password": "password123"},
			someCondition:   false,
			redisError:      false,
			expectedStatus:  http.StatusCreated,
			expectedMessage: "Signup successful",
		},
		{
			name:            "Database error during signup",
			reqBody:         map[string]string{"name": "newuser", "email": "newuser@example.com", "password": "password123"},
			someCondition:   true,
			redisError:      false,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Something went wrong, please try again later",
		},
		{
			name:            "User name already exists",
			reqBody:         map[string]string{"name": "existing_user", "email": "unique@example.com", "password": "password123"},
			someCondition:   false,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "An account with this name already exists",
		},
		{
			name:            "Email already exists",
			reqBody:         map[string]string{"name": "unique_user", "email": "existing@example.com", "password": "password123"},
			someCondition:   false,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "An account with this email already exists",
		},
		{
			name:            "Password hashing failure",
			reqBody:         map[string]string{"name": "new_user", "email": "new@example.com", "password": "password123"},
			hashPasswordErr: true,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Internal server error",
		},
		{
			name:                   "Error generating access token",
			reqBody:                map[string]string{"name": "new_user", "email": "new@example.com", "password": "password123"},
			generateAccessTokenErr: true,
			expectedStatus:         http.StatusInternalServerError,
			expectedMessage:        "Failed to generate token",
		},
		{
			name:                    "Error generating refresh token",
			reqBody:                 map[string]string{"name": "new_user", "email": "new@example.com", "password": "password123"},
			generateRefreshTokenErr: true,
			expectedStatus:          http.StatusInternalServerError,
			expectedMessage:         "Failed to generate token",
		},
		{
			name:            "Error with Redis set during signup",
			reqBody:         map[string]string{"name": "test_user", "email": "test@example.com", "password": "password123"},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to store session",
			redisError:      true,
		},
	}

	for _, tt := range tests {
		apicfg.APIConfig.RedisClient = mockRedis
		t.Run(tt.name, func(t *testing.T) {
			if tt.redisError {
				mockRedisClient.ExpectSet("refresh_token:*", "mockRefreshToken", time.Minute*60).SetErr(errors.New("Redis set error"))
			} else {
				apicfg.APIConfig.RedisClient = redis.NewClient(&redis.Options{})
			}
			patches := setupMockFunctions(tt.someCondition, tt.hashPasswordErr, tt.generateAccessTokenErr, tt.generateRefreshTokenErr)
			defer patches.Reset()

			runSignUpTest(t, apicfg, tt.reqBody, tt.expectedStatus, tt.expectedMessage)
		})
	}
}
