package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/tests/handlers/mocks"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandlerSignUp(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	testCases := []struct {
		name           string
		requestBody    map[string]string
		mockSetup      func(sqlmock.Sqlmock, redismock.ClientMock, *mocks.MockAuthHelper)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Signup Success",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}
				mockAuth.StoreRefreshTokenInRedisFn = func(r *http.Request, userID, token, provider string, duration time.Duration) error {
					data := auth.RefreshTokenData{
						Token:    token,
						Provider: provider,
					}
					jsonData, _ := json.Marshal(data)

					redisMock.ExpectSet("refresh_token:"+userID, jsonData, duration).SetVal("OK")
					return redisClient.Set(r.Context(), "refresh_token:"+userID, jsonData, duration).Err()
				}

			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Signup successful",
		},
		{
			name: "Username already exists",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "new@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "An account with this name already exists",
		},
		{
			name: "Email already exists",
			requestBody: map[string]string{
				"name":     "newuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("newuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "An account with this email already exists",
		},
		{
			name:        "Invalid JSON format",
			requestBody: nil, // will marshal to `null`
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				// no db or redis call expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request format",
		},
		{
			name: "Password hash fail",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "1",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mockAuth.HashPasswordFn = func(password string) (string, error) {
					return "", errors.New("hash error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
		{
			name: "Create user fail",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()

				mockAuth.HashPasswordFn = func(password string) (string, error) {
					return "hashed-password", nil
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later",
		},
		{
			name: "Token generate fail",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()

				mockAuth.HashPasswordFn = func(password string) (string, error) {
					return "hashed-password", nil
				}
				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "", "", errors.New("token error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to generate token",
		},
		{
			name: "Redis set fail",
			requestBody: map[string]string{
				"name":     "testuser",
				"email":    "test@example.com",
				"password": "pass12345",
			},
			mockSetup: func(mock sqlmock.Sqlmock, redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()

				mockAuth.HashPasswordFn = func(password string) (string, error) {
					return "hashed-password", nil
				}
				mockAuth.GenerateTokensFn = func(userID string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}
				mockAuth.StoreRefreshTokenInRedisFn = func(r *http.Request, userID, token, provider string, duration time.Duration) error {
					return errors.New("redis error")
				}
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
				mock.ExpectQuery(`SELECT EXISTS \(SELECT name FROM users WHERE name = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(`SELECT EXISTS \(SELECT email FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))

				mockAuth.HashPasswordFn = func(password string) (string, error) {
					return "hashed-password", nil
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to commit transaction",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare mocks
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

			apicfg := &handlers.HandlersConfig{
				APIConfig: &config.APIConfig{
					DB:          q,
					DBConn:      db,
					RedisClient: redisClient,
				},
				AuthHelper: mockAuth,
				Logger:     lg,
			}

			// Prepare request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Call handler
			apicfg.HandlerSignUp(rec, req)

			// Assert
			require.Equal(t, tc.expectedStatus, rec.Code)
			require.Contains(t, rec.Body.String(), tc.expectedBody)
			require.NoError(t, mock.ExpectationsWereMet())
			require.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
