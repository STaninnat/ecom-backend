// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
package authhandlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// auth_helper_test.go: Mock implementations and test utilities for authentication, authorization, token handling, and related service components.

// --- RefreshTokenData represents refresh token data structure ---
type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}

// --- MockAuthService is a mock implementation of AuthService for testing ---
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) SignUp(ctx context.Context, params SignUpParams) (*AuthResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) SignIn(ctx context.Context, params SignInParams) (*AuthResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) SignOut(ctx context.Context, userID string, provider string) error {
	args := m.Called(ctx, userID, provider)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, userID string, provider string, refreshToken string) (*AuthResult, error) {
	args := m.Called(ctx, userID, provider, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) HandleGoogleAuth(ctx context.Context, code string, state string) (*AuthResult, error) {
	args := m.Called(ctx, code, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) GenerateGoogleAuthURL(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

// --- MockHandlersConfig is a mock implementation of HandlersConfig for testing ---
type MockHandlersConfig struct {
	mock.Mock
}

func (m *MockHandlersConfig) LogHandlerError(ctx context.Context, operation, errorCode, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, errorCode, message, ip, userAgent, err)
}

func (m *MockHandlersConfig) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}

// --- MockCartConfig is a mock implementation of HandlersCartConfig for testing ---
type MockCartConfig struct {
	mock.Mock
}

func (m *MockCartConfig) MergeCart(ctx context.Context, r *http.Request, userID string) {
	m.Called(ctx, r, userID)
}

// --- MockAuthConfig is a mock implementation of auth config for testing ---
type mockAuthConfig struct {
	mock.Mock
}

func (m *mockAuthConfig) ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (string, *RefreshTokenData, error) {
	args := m.Called(w, r)
	return args.String(0), args.Get(1).(*RefreshTokenData), args.Error(2)
}

func (m *mockAuthConfig) SetTokensAsCookies(w http.ResponseWriter, accessToken, refreshToken string, accessTokenExpires, refreshTokenExpires time.Time) {
	m.Called(w, accessToken, refreshToken, accessTokenExpires, refreshTokenExpires)
}

// --- TestHandlersAuthConfig is a test configuration that embeds the mocks ---
type TestHandlersAuthConfig struct {
	*MockHandlersConfig
	*MockCartConfig
	Auth        *mockAuthConfig
	authService AuthService
}

func (cfg *TestHandlersAuthConfig) GetAuthService() AuthService {
	return cfg.authService
}

// --- SetupTestConfig creates and returns a TestHandlersAuthConfig with all necessary mocks for testing ---
func setupTestConfig() *TestHandlersAuthConfig {
	return &TestHandlersAuthConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		MockCartConfig:     &MockCartConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
	}
}

// --- Extend TestHandlersAuthConfig to include Auth field and HandlerSignOut method ---
func (cfg *TestHandlersAuthConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.LogHandlerError(
			ctx,
			"sign_out",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	err = cfg.GetAuthService().SignOut(ctx, userID, storedData.Provider)
	if err != nil {
		cfg.handleAuthError(w, r, err, "sign_out", ip, userAgent)
		return
	}

	// Clear cookies
	timeNow := time.Now().UTC()
	expiredTime := timeNow.Add(-1 * time.Hour)
	auth.SetTokensAsCookies(w, "", "", expiredTime, expiredTime)

	// Handle Google revoke if needed
	if storedData.Provider == "google" {
		googleRevokeURL := "https://accounts.google.com/o/oauth2/revoke?token=" + storedData.Token
		http.Redirect(w, r, googleRevokeURL, http.StatusFound)
		return
	}

	// Log success and respond
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.LogHandlerSuccess(ctxWithUserID, "sign_out", "Sign out success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Sign out successful",
	})
}

// --- Extend HandlerRefreshToken is a test handler implementation using mocked dependencies for refresh token logic ---
func (cfg *TestHandlersAuthConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token using mocked auth
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.LogHandlerError(
			ctx,
			"refresh_token",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().RefreshToken(ctx, userID, storedData.Provider, storedData.Token)
	if err != nil {
		cfg.handleAuthError(w, r, err, "refresh_token", ip, userAgent)
		return
	}

	// Set new cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := ctx // We don't have utils.ContextKeyUserID in test context
	cfg.LogHandlerSuccess(ctxWithUserID, "refresh_token", "Refresh token success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Token refreshed successfully",
	})
}

// handleAuthError handles authentication-specific errors with proper logging and responses
// It categorizes errors and provides appropriate HTTP status codes and messages
func (cfg *TestHandlersAuthConfig) handleAuthError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "name_exists", "email_exists", "user_not_found", "invalid_password":
			cfg.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "database_error", "transaction_error", "create_user_error", "hash_error", "token_generation_error", "redis_error", "commit_error", "update_user_error", "uuid_error":
			cfg.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "invalid_state", "token_exchange_error", "google_api_error", "no_refresh_token", "google_token_error", "token_expired":
			cfg.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		default:
			cfg.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// HandlerGoogleSignIn is a test handler that simulates the Google sign-in flow using mocked dependencies.
func (cfg *TestHandlersAuthConfig) HandlerGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	// Generate state and auth URL
	state := "test-state" // Mock state generation
	authURL, err := cfg.GetAuthService().GenerateGoogleAuthURL(state)
	if err != nil {
		cfg.LogHandlerError(
			r.Context(),
			"signin-google",
			"auth_url_generation_failed",
			"Error generating Google auth URL",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to initiate Google signin")
		return
	}

	// Redirect to Google
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandlerGoogleCallback is a test handler that simulates the Google OAuth callback using mocked dependencies.
func (cfg *TestHandlersAuthConfig) HandlerGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get parameters from URL
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		cfg.LogHandlerError(
			ctx,
			"callback-google",
			"missing_parameters",
			"Missing state or code parameter",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().HandleGoogleAuth(ctx, code, state)
	if err != nil {
		cfg.handleAuthError(w, r, err, "callback-google", ip, userAgent)
		return
	}

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := ctx // We don't have utils.ContextKeyUserID in test context
	cfg.LogHandlerSuccess(ctxWithUserID, "callback-google", "Google signin success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Google signin successful",
	})
}

// --- Mocks for new interfaces ---
type MockDBTx struct {
	commitFunc   func() error
	rollbackFunc func() error
}

func (m *MockDBTx) Commit() error {
	if m.commitFunc != nil {
		return m.commitFunc()
	}
	return nil
}
func (m *MockDBTx) Rollback() error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc()
	}
	return nil
}

type MockDBConn struct {
	beginTxFunc func(ctx context.Context, opts *sql.TxOptions) (DBTx, error)
}

func (m *MockDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTx, error) {
	if m.beginTxFunc != nil {
		return m.beginTxFunc(ctx, opts)
	}
	return &MockDBTx{}, nil
}

type MockDBQueries struct {
	CheckUserExistsByNameFunc         func(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmailFunc        func(ctx context.Context, email string) (bool, error)
	CreateUserFunc                    func(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmailFunc                func(ctx context.Context, email string) (database.User, error)
	UpdateUserStatusByIDFunc          func(ctx context.Context, params database.UpdateUserStatusByIDParams) error
	WithTxFunc                        func(tx DBTx) DBQueries
	CheckExistsAndGetIDByEmailFunc    func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error)
	UpdateUserSigninStatusByEmailFunc func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error
}

func (m *MockDBQueries) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return m.CheckUserExistsByNameFunc(ctx, name)
}
func (m *MockDBQueries) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.CheckUserExistsByEmailFunc(ctx, email)
}
func (m *MockDBQueries) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return m.CreateUserFunc(ctx, params)
}
func (m *MockDBQueries) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return m.GetUserByEmailFunc(ctx, email)
}
func (m *MockDBQueries) UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
	return m.UpdateUserStatusByIDFunc(ctx, params)
}
func (m *MockDBQueries) WithTx(tx DBTx) DBQueries {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}
func (m *MockDBQueries) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	return m.CheckExistsAndGetIDByEmailFunc(ctx, email)
}
func (m *MockDBQueries) UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
	return m.UpdateUserSigninStatusByEmailFunc(ctx, params)
}

// mockServiceAuthConfig is a mock implementation of the AuthConfig interface for service-level tests.
type mockServiceAuthConfig struct{}

// HashPassword is a mock implementation for password hashing in tests.
func (m *mockServiceAuthConfig) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

// GenerateTokens is a mock implementation for token generation in tests.
func (m *mockServiceAuthConfig) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	cfg := &auth.Config{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	return cfg.GenerateTokens(userID, expiresAt)
}

// StoreRefreshTokenInRedis is a mock implementation for storing refresh tokens in tests.
func (m *mockServiceAuthConfig) StoreRefreshTokenInRedis(_ context.Context, _, _, _ string, _ time.Duration) error {
	return nil
}

// GenerateAccessToken is a mock implementation for access token generation in tests.
func (m *mockServiceAuthConfig) GenerateAccessToken(_ string, _ time.Time) (string, error) {
	return "access-token", nil
}

// --- DBQueriesAdapter method forwarding tests ---
type fakeQueries struct {
	CheckUserExistsByNameFunc         func(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmailFunc        func(ctx context.Context, email string) (bool, error)
	CreateUserFunc                    func(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmailFunc                func(ctx context.Context, email string) (database.User, error)
	UpdateUserStatusByIDFunc          func(ctx context.Context, params database.UpdateUserStatusByIDParams) error
	WithTxFunc                        func(tx any) *fakeQueries
	CheckExistsAndGetIDByEmailFunc    func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error)
	UpdateUserSigninStatusByEmailFunc func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error
}

func (f *fakeQueries) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return f.CheckUserExistsByNameFunc(ctx, name)
}
func (f *fakeQueries) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return f.CheckUserExistsByEmailFunc(ctx, email)
}
func (f *fakeQueries) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return f.CreateUserFunc(ctx, params)
}
func (f *fakeQueries) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return f.GetUserByEmailFunc(ctx, email)
}
func (f *fakeQueries) UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
	return f.UpdateUserStatusByIDFunc(ctx, params)
}
func (f *fakeQueries) WithTx(tx any) *fakeQueries {
	return f.WithTxFunc(tx)
}
func (f *fakeQueries) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	return f.CheckExistsAndGetIDByEmailFunc(ctx, email)
}
func (f *fakeQueries) UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
	return f.UpdateUserSigninStatusByEmailFunc(ctx, params)
}

// For testing, define a minimal DBQueries interface and use composition instead of embedding
type testDBQueriesAdapter struct {
	*fakeQueries
}

func (a *testDBQueriesAdapter) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return a.fakeQueries.CheckUserExistsByName(ctx, name)
}
func (a *testDBQueriesAdapter) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.fakeQueries.CheckUserExistsByEmail(ctx, email)
}
func (a *testDBQueriesAdapter) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return a.fakeQueries.CreateUser(ctx, params)
}
func (a *testDBQueriesAdapter) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return a.fakeQueries.GetUserByEmail(ctx, email)
}
func (a *testDBQueriesAdapter) UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
	return a.fakeQueries.UpdateUserStatusByID(ctx, params)
}
func (a *testDBQueriesAdapter) WithTx(tx interface{}) *fakeQueries {
	return a.fakeQueries.WithTx(tx)
}
func (a *testDBQueriesAdapter) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	return a.fakeQueries.CheckExistsAndGetIDByEmail(ctx, email)
}
func (a *testDBQueriesAdapter) UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
	return a.fakeQueries.UpdateUserSigninStatusByEmail(ctx, params)
}

// --- Minimal fakeRedis for MinimalRedis interface ---
type FakeRedis struct {
	getResult string
}

func (f *FakeRedis) Del(_ context.Context, _ ...string) *redis.IntCmd {
	return redis.NewIntResult(1, nil)
}
func (f *FakeRedis) Set(_ context.Context, _ string, _ any, _ time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("OK", nil)
}
func (f *FakeRedis) Get(_ context.Context, _ string) *redis.StringCmd {
	return redis.NewStringResult(f.getResult, nil)
}

// Add other required redis.Cmdable methods as needed for your tests

// --- Test for SignOut ---
type ErrorRedis struct{}

func (e *ErrorRedis) Del(_ context.Context, _ ...string) *redis.IntCmd {
	return redis.NewIntResult(0, assert.AnError)
}
func (e *ErrorRedis) Set(_ context.Context, _ string, _ any, _ time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("", assert.AnError)
}
func (e *ErrorRedis) Get(_ context.Context, _ string) *redis.StringCmd {
	return redis.NewStringResult("", assert.AnError)
}

// --- Mocks for error cases ---
type mockAuthConfigWithTokenError struct{}

func (m *mockAuthConfigWithTokenError) HashPassword(_ string) (string, error) { return "", nil }
func (m *mockAuthConfigWithTokenError) GenerateTokens(_ string, _ time.Time) (string, string, error) {
	return "", "", assert.AnError
}
func (m *mockAuthConfigWithTokenError) StoreRefreshTokenInRedis(_ context.Context, _, _, _ string, _ time.Duration) error {
	return nil
}
func (m *mockAuthConfigWithTokenError) GenerateAccessToken(_ string, _ time.Time) (string, error) {
	return "", assert.AnError
}

type mockAuthConfigWithStoreError struct{}

func (m *mockAuthConfigWithStoreError) HashPassword(_ string) (string, error) { return "", nil }
func (m *mockAuthConfigWithStoreError) GenerateTokens(_ string, _ time.Time) (string, string, error) {
	return "access", "refresh", nil
}
func (m *mockAuthConfigWithStoreError) StoreRefreshTokenInRedis(_ context.Context, _, _, _ string, _ time.Duration) error {
	return assert.AnError
}
func (m *mockAuthConfigWithStoreError) GenerateAccessToken(_ string, _ time.Time) (string, error) {
	return "", nil
}

type mockAuthConfigWithHashError struct{}

func (m *mockAuthConfigWithHashError) HashPassword(_ string) (string, error) {
	return "", assert.AnError
}
func (m *mockAuthConfigWithHashError) GenerateTokens(_ string, _ time.Time) (string, string, error) {
	return "", "", nil
}
func (m *mockAuthConfigWithHashError) StoreRefreshTokenInRedis(_ context.Context, _, _, _ string, _ time.Duration) error {
	return nil
}
func (m *mockAuthConfigWithHashError) GenerateAccessToken(_ string, _ time.Time) (string, error) {
	return "", nil
}

// --- Mocks for OAuth ---
type mockOAuth2ExchangerWithClient struct {
	oauth2.Config
	client *http.Client
}

func (m *mockOAuth2ExchangerWithClient) Client(_ context.Context, _ *oauth2.Token) *http.Client {
	return m.client
}

func (m *mockOAuth2ExchangerWithClient) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

func (m *mockOAuth2ExchangerWithClient) AuthCodeURL(_ string, _ ...oauth2.AuthCodeOption) string {
	return ""
}

func (m *mockOAuth2ExchangerWithClient) TokenSource(_ context.Context, _ *oauth2.Token) oauth2.TokenSource {
	return nil
}

// --- Mocks for OAuth2 for HandleGoogleAuth tests ---
type mockOAuth2Config struct {
	Config        oauth2.Config
	exchangeToken *oauth2.Token
	exchangeErr   error
}

func (m *mockOAuth2Config) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return m.exchangeToken, m.exchangeErr
}
func (m *mockOAuth2Config) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	return "http://mock-oauth-url/?state=" + state
}
func (m *mockOAuth2Config) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return m.Config.TokenSource(ctx, t)
}
func (m *mockOAuth2Config) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return m.Config.Client(ctx, t)
}

// --- Mocks for MinimalRedis and OAuth2Exchanger ---
type mockRedisClient struct {
	DelFunc func(ctx context.Context, keys ...string) *redis.IntCmd
	SetFunc func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return m.DelFunc(ctx, keys...)
}
func (m *mockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return m.SetFunc(ctx, key, value, expiration)
}
func (m *mockRedisClient) Get(_ context.Context, _ string) *redis.StringCmd {
	return redis.NewStringResult("", nil)
}

type mockOAuth2Exchanger struct {
	AuthCodeURLFunc func(state string, opts ...oauth2.AuthCodeOption) string
}

func (m *mockOAuth2Exchanger) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return m.AuthCodeURLFunc(state, opts...)
}
func (m *mockOAuth2Exchanger) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}
func (m *mockOAuth2Exchanger) TokenSource(_ context.Context, _ *oauth2.Token) oauth2.TokenSource {
	return nil
}
func (m *mockOAuth2Exchanger) Client(_ context.Context, _ *oauth2.Token) *http.Client { return nil }

// --- TestMergeCartConfig is a test configuration specifically for MergeCart tests ---
type TestMergeCartConfig struct {
	*MockHandlersConfig
	Auth            *mockAuthConfig
	authService     AuthService
	CartMG          *MockCartManager
	GetGuestCart    func(ctx context.Context, sessionID string) (*models.Cart, error)
	DeleteGuestCart func(ctx context.Context, sessionID string) error
}

func (cfg *TestMergeCartConfig) GetAuthService() AuthService {
	return cfg.authService
}

func (cfg *TestMergeCartConfig) LogHandlerError(ctx context.Context, operation, errorCode, message, ip, userAgent string, err error) {
	cfg.MockHandlersConfig.LogHandlerError(ctx, operation, errorCode, message, ip, userAgent, err)
}

// MergeCart is the real MergeCart function for testing
func (cfg *TestMergeCartConfig) MergeCart(ctx context.Context, r *http.Request, userID string) {
	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		return
	}

	guestCart, err := cfg.GetGuestCart(ctx, sessionID)
	if err != nil {
		// Log error but don't fail the authentication process
		cfg.LogHandlerError(ctx, "merge_cart", "get_guest_cart_failed", "Failed to get guest cart", "", "", err)
		return
	}

	if len(guestCart.Items) == 0 {
		// No items to merge, just clean up the guest cart
		if err := cfg.DeleteGuestCart(ctx, sessionID); err != nil {
			cfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete empty guest cart", "", "", err)
		}
		return
	}

	// Merge guest cart items to user cart
	if err := cfg.CartMG.MergeGuestCartToUser(ctx, userID, guestCart.Items); err != nil {
		cfg.LogHandlerError(ctx, "merge_cart", "merge_cart_failed", "Failed to merge guest cart to user", "", "", err)
		return
	}

	// Clean up guest cart after successful merge
	if err := cfg.DeleteGuestCart(ctx, sessionID); err != nil {
		cfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete guest cart after merge", "", "", err)
	}
}

type MockCartManager struct {
	mock.Mock
}

func (m *MockCartManager) MergeGuestCartToUser(ctx context.Context, userID string, items []models.CartItem) error {
	args := m.Called(ctx, userID, items)
	return args.Error(0)
}

// --- RefreshGoogleToken Error Path and Happy Path Tests ---

type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

type mockOAuth2ExchangerWithTokenSource struct {
	oauth2.Config
	tokenSource *mockTokenSource
}

func (m *mockOAuth2ExchangerWithTokenSource) TokenSource(_ context.Context, _ *oauth2.Token) oauth2.TokenSource {
	return m.tokenSource
}

func (m *mockOAuth2ExchangerWithTokenSource) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

func (m *mockOAuth2ExchangerWithTokenSource) AuthCodeURL(_ string, _ ...oauth2.AuthCodeOption) string {
	return ""
}

func (m *mockOAuth2ExchangerWithTokenSource) Client(_ context.Context, _ *oauth2.Token) *http.Client {
	return nil
}
