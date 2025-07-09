package authhandlers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// TestAuthServiceInterface tests that the AuthService interface is properly defined
func TestAuthServiceInterface(t *testing.T) {
	// This test ensures that the AuthService interface is properly defined
	// and that all required methods are present
	var _ AuthService = (*AuthServiceImpl)(nil)
}

// TestSignUpParams tests the SignUpParams struct
func TestSignUpParams(t *testing.T) {
	params := SignUpParams{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "testuser", params.Name)
	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, "password123", params.Password)
}

// TestSignInParams tests the SignInParams struct
func TestSignInParams(t *testing.T) {
	params := SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, "password123", params.Password)
}

// TestAuthResult tests the AuthResult struct
func TestAuthResult(t *testing.T) {
	result := &AuthResult{
		UserID:       "user-id",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		IsNewUser:    true,
	}

	assert.Equal(t, "user-id", result.UserID)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, "refresh-token", result.RefreshToken)
	assert.True(t, result.IsNewUser)
}

// TestAuthError tests the AuthError struct
func TestAuthError(t *testing.T) {
	err := &handlers.AppError{
		Code:    "test_error",
		Message: "Test error message",
		Err:     nil,
	}

	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "Test error message", err.Message)
	assert.Nil(t, err.Err)
	assert.Equal(t, "Test error message", err.Error())
}

// TestAuthError_WithInnerError tests AuthError with an inner error
func TestAuthError_WithInnerError(t *testing.T) {
	innerErr := assert.AnError
	err := &handlers.AppError{
		Code:    "test_error",
		Message: "Test error message",
		Err:     innerErr,
	}

	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "Test error message", err.Message)
	assert.Equal(t, innerErr, err.Err)
	assert.Contains(t, err.Error(), "Test error message")
	assert.Contains(t, err.Error(), innerErr.Error())
}

// TestNewAuthService tests the NewAuthService function
func TestNewAuthService(t *testing.T) {
	// This test ensures that NewAuthService returns a valid AuthService
	// The actual implementation would require real dependencies
	// For now, we just test that the function exists and can be called
	assert.NotNil(t, NewAuthService)
}

// TestConstants tests that the constants are properly defined
func TestConstants(t *testing.T) {
	assert.NotZero(t, AccessTokenTTL)
	assert.NotZero(t, RefreshTokenTTL)
	assert.NotZero(t, OAuthStateTTL)
	assert.Equal(t, "local", LocalProvider)
	assert.Equal(t, "user", UserRole)
	assert.Equal(t, "refresh_token:", RefreshTokenKeyPrefix)
	assert.Equal(t, "oauth_state:", OAuthStateKeyPrefix)
	assert.Equal(t, "valid", OAuthStateValid)
}

// Mocks for new interfaces

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

// Minimal fakeRedis for MinimalRedis interface
type FakeRedis struct {
	getResult string
}

func (f *FakeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntResult(1, nil)
}
func (f *FakeRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("OK", nil)
}
func (f *FakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringResult(f.getResult, nil)
}

// Add other required redis.Cmdable methods as needed for your tests

// --- Test for SignOut ---
type ErrorRedis struct{}

func (e *ErrorRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntResult(0, assert.AnError)
}
func (e *ErrorRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("", assert.AnError)
}
func (e *ErrorRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringResult("", assert.AnError)
}

func TestAuthServiceImpl_RefreshToken(t *testing.T) {
	t.Run("local provider success", func(t *testing.T) {
		svc := &AuthServiceImpl{
			redisClient: &FakeRedis{},
			db:          nil,
			dbConn:      nil,
			auth:        &mockServiceAuthConfig{},
		}
		result, err := svc.RefreshToken(context.Background(), "user123", "local", "token")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("local provider redis error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			redisClient: &ErrorRedis{},
			db:          nil,
			dbConn:      nil,
			auth:        &mockServiceAuthConfig{},
		}
		result, err := svc.RefreshToken(context.Background(), "user123", "local", "token")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAuthServiceImpl_refreshGoogleToken_Error(t *testing.T) {
	svc := &AuthServiceImpl{
		oauth: &oauth2.Config{},
	}
	// Use an empty token, which will cause the TokenSource to fail
	result, err := svc.refreshGoogleToken(context.Background(), "user123", "", time.Now())
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestAuthServiceImpl_generateAndStoreTokens tests the generateAndStoreTokens function
func TestAuthServiceImpl_generateAndStoreTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockServiceAuthConfig{},
		}
		ctx := context.Background()
		userID := "user123"
		provider := "local"
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, provider, timeNow, true)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("token generation error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockAuthConfigWithTokenError{},
		}
		ctx := context.Background()
		userID := "user123"
		provider := "local"
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, provider, timeNow, true)
		assert.Error(t, err)
		assert.Nil(t, result)
		appErr, ok := err.(*handlers.AppError)
		assert.True(t, ok)
		assert.Equal(t, "token_generation_error", appErr.Code)
	})

	t.Run("refresh token storage error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockAuthConfigWithStoreError{},
		}
		ctx := context.Background()
		userID := "user123"
		provider := "local"
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, provider, timeNow, true)
		assert.Error(t, err)
		assert.Nil(t, result)
		appErr, ok := err.(*handlers.AppError)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

// --- Test Template for Service Methods ---
// Use mockDBQueries, mockDBConn, mockDBTx, mockServiceAuthConfig, and fakeRedis for all tests
// For custom behavior, define closures in the test setup

// Example: Success case for SignUp using the new fakeRedis
func TestAuthServiceImpl_SignUp_Success_WithMinimalRedis(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
		CreateUserFunc:             func(ctx context.Context, params database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{
		beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil },
	}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// Example: Custom error case for SignUp (e.g., duplicate email)
func TestAuthServiceImpl_SignUp_DuplicateEmail_Template(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return true, nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "pass"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, "An account with this email already exists", err.Error())
}

// Example: Success case for SignIn
func TestAuthServiceImpl_SignIn_Success_Template(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	// The hash must match what CheckPasswordHash expects
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// --- End of Template ---
// For each new test, copy the pattern above and override only the methods you need for the scenario.

// Rename mockAuthConfig to mockServiceAuthConfig for service-level tests
type mockServiceAuthConfig struct{}

func (m *mockServiceAuthConfig) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}
func (m *mockServiceAuthConfig) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	cfg := &auth.AuthConfig{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	return cfg.GenerateTokens(userID, expiresAt)
}
func (m *mockServiceAuthConfig) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	return nil
}
func (m *mockServiceAuthConfig) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return "access-token", nil
}

// Mocks for error cases

type mockAuthConfigWithTokenError struct{}

func (m *mockAuthConfigWithTokenError) HashPassword(password string) (string, error) { return "", nil }
func (m *mockAuthConfigWithTokenError) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return "", "", assert.AnError
}
func (m *mockAuthConfigWithTokenError) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	return nil
}
func (m *mockAuthConfigWithTokenError) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return "", assert.AnError
}

type mockAuthConfigWithStoreError struct{}

func (m *mockAuthConfigWithStoreError) HashPassword(password string) (string, error) { return "", nil }
func (m *mockAuthConfigWithStoreError) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return "access", "refresh", nil
}
func (m *mockAuthConfigWithStoreError) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	return assert.AnError
}
func (m *mockAuthConfigWithStoreError) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return "", nil
}

type mockAuthConfigWithHashError struct{}

func (m *mockAuthConfigWithHashError) HashPassword(password string) (string, error) {
	return "", assert.AnError
}
func (m *mockAuthConfigWithHashError) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return "", "", nil
}
func (m *mockAuthConfigWithHashError) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	return nil
}
func (m *mockAuthConfigWithHashError) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return "", nil
}

// SignUp Error Path Tests

func TestAuthServiceImpl_SignUp_CheckNameExistsError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc: func(ctx context.Context, name string) (bool, error) {
			return false, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error checking name existence")
}

func TestAuthServiceImpl_SignUp_CheckEmailExistsError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc: func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error checking email existence")
}

func TestAuthServiceImpl_SignUp_HashPasswordError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithHashError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error hashing password")
}

func TestAuthServiceImpl_SignUp_BeginTxError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) {
		return nil, assert.AnError
	}}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error starting transaction")
}

func TestAuthServiceImpl_SignUp_CreateUserError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
		CreateUserFunc:             func(ctx context.Context, params database.CreateUserParams) error { return assert.AnError },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error creating user")
}

func TestAuthServiceImpl_SignUp_TokenGenerationError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
		CreateUserFunc:             func(ctx context.Context, params database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithTokenError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error generating tokens")
}

func TestAuthServiceImpl_SignUp_StoreTokenError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
		CreateUserFunc:             func(ctx context.Context, params database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithStoreError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error storing refresh token")
}

func TestAuthServiceImpl_SignUp_CommitError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(ctx context.Context, name string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) { return false, nil },
		CreateUserFunc:             func(ctx context.Context, params database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockTx := &MockDBTx{commitFunc: func() error { return assert.AnError }}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return mockTx, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: "longenoughpassword"}
	result, err := service.SignUp(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error committing transaction")
}

// SignIn Error Path Tests

func TestAuthServiceImpl_SignIn_GetUserByEmailError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{}, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: "password"}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Invalid credentials")
}

func TestAuthServiceImpl_SignIn_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: "wrongpassword"}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Invalid credentials")
}

func TestAuthServiceImpl_SignIn_UUIDParseError(t *testing.T) {
	ctx := context.Background()
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: "invalid-uuid", Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Invalid user ID")
}

func TestAuthServiceImpl_SignIn_BeginTxError(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) {
		return nil, assert.AnError
	}}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error starting transaction")
}

func TestAuthServiceImpl_SignIn_UpdateUserStatusError(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
			return assert.AnError
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error updating user status")
}

func TestAuthServiceImpl_SignIn_TokenGenerationError(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithTokenError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error generating tokens")
}

func TestAuthServiceImpl_SignIn_StoreTokenError(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithStoreError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error storing refresh token")
}

func TestAuthServiceImpl_SignIn_CommitError(t *testing.T) {
	ctx := context.Background()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	password := "longenoughpassword"
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockTx := &MockDBTx{commitFunc: func() error { return assert.AnError }}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return mockTx, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: password}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error committing transaction")
}

func TestAuthServiceImpl_getUserInfoFromGoogle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user := &UserGoogleInfo{ID: "id", Name: "name", Email: "email@example.com"}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(user)
		}))
		defer ts.Close()
		client := ts.Client()
		svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
		result, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL, client)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.Email, result.Email)
	})

	t.Run("http error", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, assert.AnError
		})}
		svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
		_, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, "http://example.com", client)
		assert.Error(t, err)
	})

	t.Run("json decode error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer ts.Close()
		client := ts.Client()
		svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
		_, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL, client)
		assert.Error(t, err)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func Test_getUserInfoFromGoogle_withCustomURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"123","email":"test@example.com","name":"Test User"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
	client := ts.Client()
	user, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL+"/userinfo", client)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)

	// Error case: invalid JSON
	tsErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":`)) // malformed JSON
	}))
	defer tsErr.Close()

	_, err = svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, tsErr.URL, tsErr.Client())
	assert.Error(t, err)
}

func Test_getUserInfoFromGoogle_NoCustomClient(t *testing.T) {
	// Test when no custom client is provided - should use s.oauth.Client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","email":"test@example.com","name":"Test User"}`))
	}))
	defer ts.Close()

	// Mock the oauth client to return our test server
	mockOAuth := &mockOAuth2ExchangerWithClient{
		client: ts.Client(),
	}

	svc := &AuthServiceImpl{oauth: mockOAuth}
	// Test the path where no custom client is provided (uses s.oauth.Client)
	user, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
}

func Test_getUserInfoFromGoogle_DefaultClient(t *testing.T) {
	// Test when no custom client is provided - should use s.oauth.Client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","email":"default@example.com","name":"Default Client"}`))
	}))
	defer ts.Close()

	// Mock the oauth client to return our test server
	mockOAuth := &mockOAuth2ExchangerWithClient{
		client: ts.Client(),
	}

	svc := &AuthServiceImpl{oauth: mockOAuth}
	user, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "456", user.ID)
	assert.Equal(t, "default@example.com", user.Email)
	assert.Equal(t, "Default Client", user.Name)
}

type mockOAuth2ExchangerWithClient struct {
	oauth2.Config
	client *http.Client
}

func (m *mockOAuth2ExchangerWithClient) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return m.client
}

func (m *mockOAuth2ExchangerWithClient) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

func (m *mockOAuth2ExchangerWithClient) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return ""
}

func (m *mockOAuth2ExchangerWithClient) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return nil
}

func TestAuthServiceImpl_HandleGoogleAuth(t *testing.T) {
	t.Run("happy path - new user", func(t *testing.T) {
		ctx := context.Background()
		redis := &FakeRedis{getResult: "valid"}
		mockDB := &MockDBQueries{
			CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
				return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
			},
			CreateUserFunc:                    func(ctx context.Context, params database.CreateUserParams) error { return nil },
			UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error { return nil },
		}
		mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
		mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
		mockAuth := &mockServiceAuthConfig{}
		mockOAuth := &mockOAuth2Config{
			Config:        oauth2.Config{},
			exchangeToken: &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"},
		}
		ts := &testAuthServiceImpl{
			AuthServiceImpl: AuthServiceImpl{
				db:          mockDB,
				dbConn:      mockConn,
				auth:        mockAuth,
				redisClient: redis,
				oauth:       mockOAuth,
			},
			getUserInfoFromGoogleFunc: func(token *oauth2.Token, userInfoURL string, clientOpt ...*http.Client) (*UserGoogleInfo, error) {
				return &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}, nil
			},
		}
		result, err := ts.HandleGoogleAuth(ctx, "code", "state")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// The returned UserID will be a generated UUID, so just check for non-empty
		assert.NotEmpty(t, result.UserID)
	})

	t.Run("invalid state", func(t *testing.T) {
		ctx := context.Background()
		redis := &FakeRedis{getResult: "invalid"}
		ts := &testAuthServiceImpl{
			AuthServiceImpl: AuthServiceImpl{redisClient: redis, oauth: &oauth2.Config{}},
		}
		result, err := ts.HandleGoogleAuth(ctx, "code", "state")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("token exchange failure", func(t *testing.T) {
		ctx := context.Background()
		redis := &FakeRedis{getResult: "valid"}
		mockOAuth := &mockOAuth2Config{Config: oauth2.Config{}, exchangeErr: assert.AnError}
		ts := &testAuthServiceImpl{
			AuthServiceImpl: AuthServiceImpl{redisClient: redis, oauth: mockOAuth},
		}
		result, err := ts.HandleGoogleAuth(ctx, "code", "state")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// --- Mocks for OAuth2 for HandleGoogleAuth tests ---
type mockOAuth2Config struct {
	Config        oauth2.Config
	exchangeToken *oauth2.Token
	exchangeErr   error
}

func (m *mockOAuth2Config) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return m.exchangeToken, m.exchangeErr
}
func (m *mockOAuth2Config) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return "http://mock-oauth-url/?state=" + state
}
func (m *mockOAuth2Config) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return m.Config.TokenSource(ctx, t)
}
func (m *mockOAuth2Config) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return m.Config.Client(ctx, t)
}

// Test double for authServiceImpl to override getUserInfoFromGoogle
type testAuthServiceImpl struct {
	AuthServiceImpl
	getUserInfoFromGoogleFunc func(token *oauth2.Token, userInfoURL string, clientOpt ...*http.Client) (*UserGoogleInfo, error)
}

// func (s *testAuthServiceImpl) getUserInfoFromGoogle(token *oauth2.Token, userInfoURL string, clientOpt ...*http.Client) (*UserGoogleInfo, error) {
// 	if s.getUserInfoFromGoogleFunc != nil {
// 		return s.getUserInfoFromGoogleFunc(token, userInfoURL, clientOpt...)
// 	}
// 	return s.AuthServiceImpl.getUserInfoFromGoogle(token, userInfoURL, clientOpt...)
// }

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
func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringResult("", nil)
}

type mockOAuth2Exchanger struct {
	AuthCodeURLFunc func(state string, opts ...oauth2.AuthCodeOption) string
}

func (m *mockOAuth2Exchanger) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return m.AuthCodeURLFunc(state, opts...)
}
func (m *mockOAuth2Exchanger) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}
func (m *mockOAuth2Exchanger) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return nil
}
func (m *mockOAuth2Exchanger) Client(ctx context.Context, t *oauth2.Token) *http.Client { return nil }

func TestAuthServiceImpl_SignOut(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			DelFunc: func(ctx context.Context, keys ...string) *redis.IntCmd {
				assert.Contains(t, keys[0], "refresh_token:")
				return redis.NewIntResult(1, nil)
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis}
		err := svc.SignOut(context.Background(), "user123", "local")
		assert.NoError(t, err)
	})
	t.Run("redis error", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			DelFunc: func(ctx context.Context, keys ...string) *redis.IntCmd {
				return redis.NewIntResult(0, assert.AnError)
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis}
		err := svc.SignOut(context.Background(), "user123", "local")
		assert.Error(t, err)
		appErr, ok := err.(*handlers.AppError)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

func TestAuthServiceImpl_GenerateGoogleAuthURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			SetFunc: func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
				assert.Contains(t, key, "oauth_state:")
				assert.Equal(t, "valid", value)
				return redis.NewStatusResult("OK", nil)
			},
		}
		mockOAuth := &mockOAuth2Exchanger{
			AuthCodeURLFunc: func(state string, opts ...oauth2.AuthCodeOption) string {
				return "https://accounts.google.com/o/oauth2/auth?state=" + state
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis, oauth: mockOAuth}
		url, err := svc.GenerateGoogleAuthURL("xyz123")
		assert.NoError(t, err)
		assert.Contains(t, url, "state=xyz123")
	})
	t.Run("redis error", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			SetFunc: func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
				return redis.NewStatusResult("", assert.AnError)
			},
		}
		mockOAuth := &mockOAuth2Exchanger{
			AuthCodeURLFunc: func(state string, opts ...oauth2.AuthCodeOption) string {
				return ""
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis, oauth: mockOAuth}
		url, err := svc.GenerateGoogleAuthURL("failstate")
		assert.Error(t, err)
		assert.Empty(t, url)
		appErr, ok := err.(*handlers.AppError)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

// --- MergeCart tests ---
// TestMergeCartConfig is a test configuration specifically for MergeCart tests
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

func TestMergeCart_NoSessionID(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart:       func(ctx context.Context, sessionID string) (*models.Cart, error) { return nil, nil },
		DeleteGuestCart:    func(ctx context.Context, sessionID string) error { return nil },
	}

	// Create request without session ID
	req := httptest.NewRequest("POST", "/signin", nil)
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify no methods were called since session ID is empty
	cfg.MockHandlersConfig.AssertNotCalled(t, "LogHandlerError")
}

func TestMergeCart_GetGuestCartError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return nil, errors.New("redis error")
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return nil },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "get_guest_cart_failed", "Failed to get guest cart", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestMergeCart_EmptyGuestCart(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return &models.Cart{Items: []models.CartItem{}}, nil
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return nil },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertNotCalled(t, "LogHandlerError")
}

func TestMergeCart_EmptyGuestCartDeleteError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return &models.Cart{Items: []models.CartItem{}}, nil
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return errors.New("delete error") },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete empty guest cart", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
}

func TestMergeCart_MergeError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return nil },
	}

	// Mock CartMG to fail
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, "user123", mock.Anything).Return(errors.New("merge error"))

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "merge_cart_failed", "Failed to merge guest cart to user", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
	cfg.CartMG.AssertExpectations(t)
}

func TestMergeCart_Success(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return nil },
	}

	// Mock CartMG to succeed
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, "user123", mock.Anything).Return(nil)

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertNotCalled(t, "LogHandlerError")
	cfg.CartMG.AssertExpectations(t)
}

func TestMergeCart_SuccessWithDeleteError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(ctx context.Context, sessionID string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(ctx context.Context, sessionID string) error { return errors.New("delete error") },
	}

	// Mock CartMG to succeed
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, "user123", mock.Anything).Return(nil)

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete guest cart after merge", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, "user123")

	// Verify expectations
	cfg.MockHandlersConfig.AssertExpectations(t)
	cfg.CartMG.AssertExpectations(t)
}

// handleGoogleUserAuth Error Path and Happy Path Tests

func TestAuthServiceImpl_handleGoogleUserAuth_DBError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error checking user existence")
}

func TestAuthServiceImpl_handleGoogleUserAuth_BeginTxError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return nil, assert.AnError }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error starting transaction")
}

func TestAuthServiceImpl_handleGoogleUserAuth_CreateUserError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
		},
		CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) error {
			return assert.AnError
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error creating user")
}

func TestAuthServiceImpl_handleGoogleUserAuth_GenerateAccessTokenError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithTokenError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error generating access token")
}

func TestAuthServiceImpl_handleGoogleUserAuth_NoRefreshToken(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: ""}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "No refresh token provided by Google")
}

func TestAuthServiceImpl_handleGoogleUserAuth_StoreRefreshTokenError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error { return nil },
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockAuthConfigWithStoreError{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error storing refresh token")
}

func TestAuthServiceImpl_handleGoogleUserAuth_UpdateUserSigninStatusError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
			return assert.AnError
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error updating user status")
}

func TestAuthServiceImpl_handleGoogleUserAuth_CommitError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
			return nil
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockTx := &MockDBTx{commitFunc: func() error { return assert.AnError }}
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return mockTx, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Error committing transaction")
}

func TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_NewUser(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
		},
		CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) error {
			return nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
			return nil
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.IsNewUser)
}

func TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_ExistingUser(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
			return nil
		},
	}
	mockDB.WithTxFunc = func(tx DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(ctx context.Context, opts *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
	token := &oauth2.Token{RefreshToken: "refresh"}
	result, err := service.handleGoogleUserAuth(ctx, user, token)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsNewUser)
}

// refreshGoogleToken Error Path and Happy Path Tests

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

func (m *mockOAuth2ExchangerWithTokenSource) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return m.tokenSource
}

func (m *mockOAuth2ExchangerWithTokenSource) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

func (m *mockOAuth2ExchangerWithTokenSource) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return ""
}

func (m *mockOAuth2ExchangerWithTokenSource) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return nil
}

func TestAuthServiceImpl_refreshGoogleToken_Success(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-id"
	refreshToken := "test-refresh-token"
	timeNow := time.Now().UTC()

	// Mock successful token refresh
	mockToken := &oauth2.Token{
		AccessToken: "new-access-token",
		Expiry:      timeNow.Add(time.Hour),
	}
	mockTokenSource := &mockTokenSource{
		token: mockToken,
		err:   nil,
	}
	mockOAuth := &mockOAuth2ExchangerWithTokenSource{
		tokenSource: mockTokenSource,
	}

	service := &AuthServiceImpl{
		oauth: mockOAuth,
	}

	result, err := service.refreshGoogleToken(ctx, userID, refreshToken, timeNow)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, userID, result.UserID)
	require.Equal(t, mockToken.AccessToken, result.AccessToken)
	require.Equal(t, refreshToken, result.RefreshToken)
	require.Equal(t, mockToken.Expiry, result.AccessTokenExpires)
	require.Equal(t, timeNow.Add(RefreshTokenTTL), result.RefreshTokenExpires)
	require.False(t, result.IsNewUser)
}

func TestAuthServiceImpl_refreshGoogleToken_TokenError(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-id"
	refreshToken := "test-refresh-token"
	timeNow := time.Now().UTC()

	// Mock failed token refresh
	mockTokenSource := &mockTokenSource{
		token: nil,
		err:   assert.AnError,
	}
	mockOAuth := &mockOAuth2ExchangerWithTokenSource{
		tokenSource: mockTokenSource,
	}

	service := &AuthServiceImpl{
		oauth: mockOAuth,
	}

	result, err := service.refreshGoogleToken(ctx, userID, refreshToken, timeNow)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Failed to refresh Google token")
}

func TestRealMergeCart_NoSessionID(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	realAuthConfig := &auth.AuthConfig{}

	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Auth: realAuthConfig,
		},
		Logger: mockHandlersConfig,
	}

	// Create request without session ID
	req := httptest.NewRequest("POST", "/signin", nil)
	ctx := context.Background()

	// Execute - should return early without calling any methods
	cfg.MergeCart(ctx, req, "user123")

	// Verify no error logging was called since session ID is empty
	mockHandlersConfig.AssertNotCalled(t, "LogHandlerError")
}
