package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/tests/handlers/mocks"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandlerSignOut(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	mockAuth := &mocks.MockAuthHelper{}
	userID := uuid.New()
	refreshToken := "refresh-token"

	type testCase struct {
		name           string
		provider       string
		setupMock      func()
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}

	tests := []testCase{
		{
			name:     "SignOut success with local provider",
			provider: "local",
			setupMock: func() {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "local"}, nil
				}
				redisMock.ExpectDel("refresh_token:" + userID.String()).SetVal(1)
			},
			expectedCode: http.StatusOK,
			expectedBody: "Sign out successful",
		},
		{
			name:     "SignOut with google provider redirects",
			provider: "google",
			setupMock: func() {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "google"}, nil
				}
			},
			expectedCode:   http.StatusFound,
			expectedHeader: "https://accounts.google.com/o/oauth2/revoke?token=" + refreshToken,
		},
		{
			name:     "SignOut failed due to token mismatch",
			provider: "local",
			setupMock: func() {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return uuid.Nil, nil, fmt.Errorf("Invalid token")
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "Invalid token",
		},
		{
			name:     "SignOut failed due to Redis deletion failure",
			provider: "local",
			setupMock: func() {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "local"}, nil
				}
				redisMock.ExpectDel("refresh_token:" + userID.String()).SetErr(fmt.Errorf("Redis deletion error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Failed to logout",
		},
		{
			name:     "SignOut failed due to token validation failure",
			provider: "local",
			setupMock: func() {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return uuid.Nil, nil, fmt.Errorf("Invalid token")
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "Invalid token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			apicfg := &handlers.HandlersConfig{
				APIConfig: &config.APIConfig{
					RedisClient: redisClient,
				},
				AuthHelper: mockAuth,
			}

			req := httptest.NewRequest(http.MethodPost, "/signout", nil)
			rec := httptest.NewRecorder()

			apicfg.HandlerSignOut(rec, req)

			require.Equal(t, tc.expectedCode, rec.Code)

			if tc.expectedCode == http.StatusOK {
				require.Contains(t, rec.Body.String(), tc.expectedBody)
			} else if tc.expectedCode == http.StatusFound {
				require.Equal(t, tc.expectedHeader, rec.Header().Get("Location"))
			}

			require.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
