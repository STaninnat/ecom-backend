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
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func runSignUpTest(t *testing.T, apicfg *handlers.HandlersConfig, reqBody map[string]string, expectedStatus int, expectedMessage string) {
	reqJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/signup", bytes.NewReader(reqJSON))
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

func TestHandlerSignUp(t *testing.T) {
	var someCondition bool
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
			return "mockAccessToken", nil
		}).
		ApplyFunc((*auth.AuthConfig).GenerateRefreshToken, func(_ *auth.AuthConfig) (string, error) {
			return "mockRefreshToken", nil
		})

	defer patches.Reset()
	mockRedis := redis.NewClient(&redis.Options{})

	apicfg := &handlers.HandlersConfig{
		APIConfig: &config.APIConfig{
			DB:          &database.Queries{},
			RedisClient: mockRedis,
			JWTSecret:   "test-secret",
		},
		Auth: &auth.AuthConfig{},
	}

	// someCondition = false
	// runSignUpTest(t, apicfg, map[string]string{
	// 	"name":     "new_user",
	// 	"email":    "new@example.com",
	// 	"password": "password123",
	// }, http.StatusCreated, "Signup successful")

	someCondition = true
	runSignUpTest(t, apicfg, map[string]string{
		"name":     "newuser",
		"email":    "newuser@example.com",
		"password": "password123",
	}, http.StatusInternalServerError, "Something went wrong, please try again later")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "existing_user",
		"email":    "unique@example.com",
		"password": "password123",
	}, http.StatusBadRequest, "An account with this name already exists")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "unique_user",
		"email":    "existing@example.com",
		"password": "password123",
	}, http.StatusBadRequest, "An account with this email already exists")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "",
		"email":    "test@example.com",
		"password": "password123",
	}, http.StatusBadRequest, "Invalid input")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "testuser",
		"email":    "",
		"password": "password123",
	}, http.StatusBadRequest, "Invalid input")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "testuser",
		"email":    "test@example.com",
		"password": "",
	}, http.StatusBadRequest, "Invalid input")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "testuser",
		"email":    "invalid-email",
		"password": "password123",
	}, http.StatusBadRequest, "Invalid email format")

	runSignUpTest(t, apicfg, map[string]string{
		"name":     "invalid name with space!",
		"email":    "valid@example.com",
		"password": "password123",
	}, http.StatusBadRequest, "Invalid username format")

}
