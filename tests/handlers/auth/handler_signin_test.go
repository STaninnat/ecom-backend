package authhandlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/tests/handlers/mocks"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandlerSignIn(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	testCases := []struct {
		name           string
		requestBody    map[string]string
		mockSetup      func(sqlmock.Sqlmock, redismock.ClientMock, *mocks.MockAuthHelper)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Signin Success",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				now := time.Now().UTC()
				userID := uuid.New().String()

				// Mock DB for GetUserByEmail
				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"name",
						"email",
						"password",
						"provider",
						"provider_id",
						"phone",
						"address",
						"role",
						"created_at",
						"updated_at",
					}).AddRow(
						userID,
						"testuser",
						"test@example.com",
						sql.NullString{String: "hashed-pass", Valid: true},
						"local",
						sql.NullString{Valid: false},
						sql.NullString{String: "", Valid: false},
						sql.NullString{String: "", Valid: false},
						"user",
						now,
						now,
					))

				// Mock password check
				mockAuth.CheckPasswordHashFn = func(pw, hash string) error {
					return nil
				}

				// Mock token generation
				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}

				// Expect Begin
				mock.ExpectBegin()

				// Mock DB update
				mock.ExpectExec(`UPDATE users SET provider = \$2, updated_at = \$3 WHERE id = \$1`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Redis expect
				mockAuth.StoreRefreshTokenInRedisFn = func(r *http.Request, userID, token, provider string, duration time.Duration) error {
					data := auth.RefreshTokenData{
						Token:    token,
						Provider: provider,
					}
					jsonData, _ := json.Marshal(data)

					redisMock.ExpectSet("refresh_token:"+userID, jsonData, duration).SetVal("OK")
					return redisClient.Set(r.Context(), "refresh_token:"+userID, jsonData, duration).Err()
				}

				// Expect Commit
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Signin successful",
		},
		{
			name: "Invalid Email",
			requestBody: map[string]string{
				"email":    "invalid@example.com",
				"password": "pass1234",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("invalid@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "name", "email", "password", "provider", "provider_id", "phone", "address", "role", "created_at", "updated_at",
					}))

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid credentials",
		},
		{
			name: "Invalid Password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpass",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				now := time.Now().UTC()
				userID := uuid.New().String()

				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"name",
						"email",
						"password",
						"provider",
						"provider_id",
						"phone",
						"address",
						"role",
						"created_at",
						"updated_at",
					}).AddRow(
						userID,
						"testuser",
						"test@example.com",
						sql.NullString{String: "hashed-pass", Valid: true},
						"local",
						sql.NullString{Valid: false},
						sql.NullString{String: "", Valid: false},
						sql.NullString{String: "", Valid: false},
						"user",
						now,
						now,
					))

				mockAuth.CheckPasswordHashFn = func(pw, hash string) error {
					return errors.New("password mismatch")
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid credentials",
		},
		{
			name: "Token Generation Failure",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "pass1234",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				now := time.Now().UTC()
				userID := uuid.New().String()

				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"name",
						"email",
						"password",
						"provider",
						"provider_id",
						"phone",
						"address",
						"role",
						"created_at",
						"updated_at",
					}).AddRow(
						userID,
						"testuser",
						"test@example.com",
						sql.NullString{String: "hashed-pass", Valid: true},
						"local",
						sql.NullString{Valid: false},
						sql.NullString{String: "", Valid: false},
						sql.NullString{String: "", Valid: false},
						"user",
						now,
						now,
					))

				mockAuth.CheckPasswordHashFn = func(pw, hash string) error {
					return nil
				}

				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "", "", fmt.Errorf("token generation error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to generate token",
		},
		{
			name: "Redis Failure",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "pass1234",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				now := time.Now().UTC()
				userID := uuid.New().String()

				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"name",
						"email",
						"password",
						"provider",
						"provider_id",
						"phone",
						"address",
						"role",
						"created_at",
						"updated_at",
					}).AddRow(
						userID,
						"testuser",
						"test@example.com",
						sql.NullString{String: "hashed-pass", Valid: true},
						"local",
						sql.NullString{Valid: false},
						sql.NullString{String: "", Valid: false},
						sql.NullString{String: "", Valid: false},
						"user",
						now,
						now,
					))

				mockAuth.CheckPasswordHashFn = func(pw, hash string) error {
					return nil
				}

				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}

				mock.ExpectBegin()

				mock.ExpectExec(`UPDATE users SET provider = \$2, updated_at = \$3 WHERE id = \$1`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockAuth.StoreRefreshTokenInRedisFn = func(r *http.Request, userID, token, provider string, duration time.Duration) error {
					return fmt.Errorf("Redis error")
				}

				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to store session",
		},
		{
			name: "Commit transaction fail",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				now := time.Now().UTC()
				userID := uuid.New().String()

				// Mock DB for GetUserByEmail
				mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1 LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"name",
						"email",
						"password",
						"provider",
						"provider_id",
						"phone",
						"address",
						"role",
						"created_at",
						"updated_at",
					}).AddRow(
						userID,
						"testuser",
						"test@example.com",
						sql.NullString{String: "hashed-pass", Valid: true},
						"local",
						sql.NullString{Valid: false},
						sql.NullString{String: "", Valid: false},
						sql.NullString{String: "", Valid: false},
						"user",
						now,
						now,
					))
				mockAuth.CheckPasswordHashFn = func(pw, hash string) error {
					return nil
				}

				// Mock token generation
				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE users SET provider = \$2, updated_at = \$3 WHERE id = \$1`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))

			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to commit transaction",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mockAuth := &mocks.MockAuthHelper{}
			tc.mockSetup(mock, redisMock, mockAuth)

			q := database.New(db)

			buf := &bytes.Buffer{}
			lg := logrus.New()
			lg.SetFormatter(&logrus.JSONFormatter{})
			lg.SetOutput(buf)

			apicfg := &authhandlers.HandlersAuthConfig{
				HandlersConfig: &handlers.HandlersConfig{
					APIConfig: &config.APIConfig{
						DB:          q,
						DBConn:      db,
						RedisClient: redisClient,
					},
					AuthHelper: mockAuth,
					Logger:     lg,
				},
			}

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			apicfg.HandlerSignIn(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			require.Contains(t, rec.Body.String(), tc.expectedBody)
			require.NoError(t, mock.ExpectationsWereMet())
			require.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
