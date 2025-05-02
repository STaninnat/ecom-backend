package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth_handler"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/tests/handlers/mocks"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func TestRefreshGoogleAccessToken(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		mockToken    *oauth2.Token
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "valid token",
			refreshToken: "valid_refresh",
			mockToken:    &oauth2.Token{AccessToken: "access_123"},
			mockErr:      nil,
			wantErr:      false,
		},
		{
			name:         "token error",
			refreshToken: "invalid_refresh",
			mockToken:    nil,
			mockErr:      errors.New("token fetch failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apicfg := &authhandlers.HandlersAuthConfig{
				HandlersConfig: &handlers.HandlersConfig{
					CustomTokenSource: func(ctx context.Context, refreshToken string) oauth2.TokenSource {
						if refreshToken != tt.refreshToken {
							t.Errorf("unexpected refreshToken: got %s, want %s", refreshToken, tt.refreshToken)
						}
						return &mockTokenSource{
							token: tt.mockToken,
							err:   tt.mockErr,
						}
					},
				},
			}

			req, _ := http.NewRequest("GET", "/", nil)
			token, err := apicfg.RefreshGoogleAccessToken(req, tt.refreshToken)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}

			if err == nil && token.AccessToken != tt.mockToken.AccessToken {
				t.Errorf("got token = %v, want = %v", token.AccessToken, tt.mockToken.AccessToken)
			}
		})
	}
}

func TestHandlerRefreshToken(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	userID := uuid.New()
	refreshToken := "test-refresh-token"

	testCases := []struct {
		name         string
		provider     string
		setupMock    func(redismock.ClientMock, *mocks.MockAuthHelper, *authhandlers.HandlersAuthConfig)
		expectedCode int
		expectedBody string
	}{
		{
			name:     "Google provider - success",
			provider: "google",
			setupMock: func(redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper, apicfg *authhandlers.HandlersAuthConfig) {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "google"}, nil
				}
				apicfg.CustomTokenSource = func(ctx context.Context, refreshToken string) oauth2.TokenSource {
					return &mockTokenSource{
						token: &oauth2.Token{
							AccessToken: "new-access-token",
							Expiry:      time.Now().Add(30 * time.Minute),
						},
						err: nil,
					}
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: "Token refreshed successful",
		},
		{
			name:     "Google provider - refresh fail",
			provider: "google",
			setupMock: func(redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper, apicfg *authhandlers.HandlersAuthConfig) {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "google"}, nil
				}
				apicfg.CustomTokenSource = func(ctx context.Context, refreshToken string) oauth2.TokenSource {
					return &mockTokenSource{
						token: nil,
						err:   fmt.Errorf("refresh failed"),
					}
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "Failed to refresh Google token",
		},
		{
			name:     "Local provider - success",
			provider: "local",
			setupMock: func(redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper, apicfg *authhandlers.HandlersAuthConfig) {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "local"}, nil
				}
				mockAuth.GenerateTokensFn = func(uid string, expiresAt time.Time) (string, string, error) {
					return "access-token", "refresh-token", nil
				}
				mockAuth.StoreRefreshTokenInRedisFn = func(r *http.Request, uid, token, provider string, ttl time.Duration) error {
					return nil
				}
				redisMock.ExpectDel("refresh_token:" + userID.String()).SetVal(1)
			},
			expectedCode: http.StatusOK,
			expectedBody: "Token refreshed successful",
		},
		{
			name:     "Local provider - redis delete fail",
			provider: "local",
			setupMock: func(redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper, apicfg *authhandlers.HandlersAuthConfig) {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return userID, &auth.RefreshTokenData{Token: refreshToken, Provider: "local"}, nil
				}
				redisMock.ExpectDel("refresh_token:" + userID.String()).SetErr(fmt.Errorf("redis error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Failed to remove refresh token from Redis",
		},
		{
			name:     "Validation failed",
			provider: "any",
			setupMock: func(redisMock redismock.ClientMock, mockAuth *mocks.MockAuthHelper, apicfg *authhandlers.HandlersAuthConfig) {
				mockAuth.ValidateCookieRefreshTokenDataFn = func(w http.ResponseWriter, r *http.Request) (uuid.UUID, *auth.RefreshTokenData, error) {
					return uuid.Nil, nil, fmt.Errorf("invalid cookie")
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "invalid cookie",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuth := &mocks.MockAuthHelper{}

			buf := &bytes.Buffer{}
			lg := logrus.New()
			lg.SetFormatter(&logrus.JSONFormatter{})
			lg.SetOutput(buf)

			apicfg := &authhandlers.HandlersAuthConfig{
				HandlersConfig: &handlers.HandlersConfig{
					APIConfig: &config.APIConfig{
						RedisClient: redisClient,
					},
					AuthHelper: mockAuth,
					Logger:     lg,
				},
			}

			tc.setupMock(redisMock, mockAuth, apicfg)

			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			rec := httptest.NewRecorder()

			apicfg.HandlerRefreshToken(rec, req)

			require.Equal(t, tc.expectedCode, rec.Code)
			require.Contains(t, rec.Body.String(), tc.expectedBody)
			require.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
