package authhandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/stretchr/testify/mock"
)

// RefreshTokenData represents refresh token data structure
type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}

// MockAuthService is a mock implementation of AuthService for testing
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

// MockHandlersConfig is a mock implementation of HandlersConfig for testing
type MockHandlersConfig struct {
	mock.Mock
}

func (m *MockHandlersConfig) LogHandlerError(ctx context.Context, operation, errorCode, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, errorCode, message, ip, userAgent, err)
}

func (m *MockHandlersConfig) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}

// MockCartConfig is a mock implementation of HandlersCartConfig for testing
type MockCartConfig struct {
	mock.Mock
}

func (m *MockCartConfig) MergeCart(ctx context.Context, r *http.Request, userID string) {
	m.Called(ctx, r, userID)
}

// mockAuthConfig is a mock implementation of auth config for testing
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

// TestHandlersAuthConfig is a test configuration that embeds the mocks
type TestHandlersAuthConfig struct {
	*MockHandlersConfig
	*MockCartConfig
	Auth        *mockAuthConfig
	authService AuthService
}

func (cfg *TestHandlersAuthConfig) GetAuthService() AuthService {
	return cfg.authService
}

// handleAuthError handles authentication-specific errors with proper logging and responses
// It categorizes errors and provides appropriate HTTP status codes and messages
func (cfg *TestHandlersAuthConfig) handleAuthError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if appErr, ok := err.(*handlers.AppError); ok {
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
