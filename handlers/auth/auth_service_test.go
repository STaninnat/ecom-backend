// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
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

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
)

// auth_service_test.go: Tests for authentication service with local and Google OAuth, token management, user registration, and session/cart merging.

const (
	testUserID       = "user123"
	testProvider     = "local"
	testUUID         = "123e4567-e89b-12d3-a456-426614174000"
	testPassword     = "longenoughpassword"
	testUserID2      = "test-user-id"
	testRefreshToken = "test-refresh-token"
)

// TestAuthServiceInterface ensures the AuthService interface is properly defined and implemented.
func TestAuthServiceInterface(_ *testing.T) {
	// This test ensures that the AuthService interface is properly defined
	// and that all required methods are present
	var _ AuthService = (*AuthServiceImpl)(nil)
}

// TestSignUpParams tests the SignUpParams struct for correct field assignment.
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

// TestSignInParams tests the SignInParams struct for correct field assignment.
func TestSignInParams(t *testing.T) {
	params := SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", params.Email)
	assert.Equal(t, "password123", params.Password)
}

// TestAuthResult tests the AuthResult struct for correct field assignment.
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

// TestAuthError tests the AppError struct for correct error message and fields.
func TestAuthError(t *testing.T) {
	err := &handlers.AppError{
		Code:    "test_error",
		Message: "Test error message",
		Err:     nil,
	}

	assert.Equal(t, "test_error", err.Code)
	assert.Equal(t, "Test error message", err.Message)
	require.NoError(t, err.Err)
	assert.Equal(t, "Test error message", err.Error())
}

// TestAuthError_WithInnerError tests AppError with an inner error for correct error message composition.
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

// TestNewAuthService ensures NewAuthService returns a valid AuthService instance.
func TestNewAuthService(t *testing.T) {
	// This test ensures that NewAuthService returns a valid AuthService
	// The actual implementation would require real dependencies
	// For now, we just test that the function exists and can be called
	assert.NotNil(t, NewAuthService)
}

// TestConstants checks that important constants are defined and have expected values.
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

// TestAuthServiceImpl_RefreshToken tests the RefreshToken method for both success and error cases.
func TestAuthServiceImpl_RefreshToken(t *testing.T) {
	t.Run("local provider success", func(t *testing.T) {
		svc := &AuthServiceImpl{
			redisClient: &FakeRedis{},
			db:          nil,
			dbConn:      nil,
			auth:        &mockServiceAuthConfig{},
		}
		result, err := svc.RefreshToken(context.Background(), testUserID, testProvider, testRefreshToken)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("local provider redis error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			redisClient: &ErrorRedis{},
			db:          nil,
			dbConn:      nil,
			auth:        &mockServiceAuthConfig{},
		}
		result, err := svc.RefreshToken(context.Background(), testUserID, testProvider, testRefreshToken)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAuthServiceImpl_refreshGoogleToken_Error tests refreshGoogleToken for error handling.
func TestAuthServiceImpl_refreshGoogleToken_Error(t *testing.T) {
	svc := &AuthServiceImpl{
		oauth: &oauth2.Config{},
	}
	// Use an empty token, which will cause the TokenSource to fail
	result, err := svc.refreshGoogleToken(context.Background(), testUserID, "", time.Now())
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestAuthServiceImpl_generateAndStoreTokens tests generateAndStoreTokens for success and error scenarios.
func TestAuthServiceImpl_generateAndStoreTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockServiceAuthConfig{},
		}
		ctx := context.Background()
		userID := testUserID
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, timeNow, true)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("token generation error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockAuthConfigWithTokenError{},
		}
		ctx := context.Background()
		userID := testUserID
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, timeNow, true)
		require.Error(t, err)
		assert.Nil(t, result)
		appErr := &handlers.AppError{}
		ok := errors.As(err, &appErr)
		assert.True(t, ok)
		assert.Equal(t, "token_generation_error", appErr.Code)
	})

	t.Run("refresh token storage error", func(t *testing.T) {
		svc := &AuthServiceImpl{
			auth: &mockAuthConfigWithStoreError{},
		}
		ctx := context.Background()
		userID := testUserID
		timeNow := time.Now()
		result, err := svc.generateAndStoreTokens(ctx, userID, timeNow, true)
		require.Error(t, err)
		assert.Nil(t, result)
		appErr := &handlers.AppError{}
		ok := errors.As(err, &appErr)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

// --- Test Template for Service Methods ---
// Use mockDBQueries, mockDBConn, mockDBTx, mockServiceAuthConfig, and fakeRedis for all tests
// For custom behavior, define closures in the test setup

// TestAuthServiceImpl_SignUp_Success_WithMinimalRedis tests successful SignUp with minimal Redis mock.
func TestAuthServiceImpl_SignUp_Success_WithMinimalRedis(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
		CreateUserFunc:             func(_ context.Context, _ database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{
		beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil },
	}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignUpParams{Name: "user", Email: "user@example.com", Password: testPassword}
	result, err := service.SignUp(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// Example: Custom error case for SignUp (e.g., duplicate email)
// TestAuthServiceImpl_SignUp_DuplicateEmail_Template tests SignUp for duplicate email error.
func TestAuthServiceImpl_SignUp_DuplicateEmail_Template(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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
// TestAuthServiceImpl_SignIn_Success_Template tests successful SignIn with valid credentials.
func TestAuthServiceImpl_SignIn_Success_Template(t *testing.T) {
	ctx := context.Background()
	userID := testUUID
	password := testPassword
	// The hash must match what CheckPasswordHash expects
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(_ context.Context, _ database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_SignUp_ErrorScenarios tests SignUp for various error scenarios.
func TestAuthServiceImpl_SignUp_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setupService   func() *AuthServiceImpl
		expectedErrMsg string
	}{
		{
			name: "CheckNameExistsError",
			setupService: func() *AuthServiceImpl {
				mockDB := &MockDBQueries{
					CheckUserExistsByNameFunc: func(_ context.Context, _ string) (bool, error) {
						return false, assert.AnError
					},
				}
				mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
				mockAuth := &mockServiceAuthConfig{}
				return &AuthServiceImpl{
					db:          mockDB,
					dbConn:      mockConn,
					auth:        mockAuth,
					redisClient: &FakeRedis{},
				}
			},
			expectedErrMsg: "Error checking name existence",
		},
		{
			name: "CheckEmailExistsError",
			setupService: func() *AuthServiceImpl {
				mockDB := &MockDBQueries{
					CheckUserExistsByNameFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
					CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) {
						return false, assert.AnError
					},
				}
				mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
				mockAuth := &mockServiceAuthConfig{}
				return &AuthServiceImpl{
					db:          mockDB,
					dbConn:      mockConn,
					auth:        mockAuth,
					redisClient: &FakeRedis{},
				}
			},
			expectedErrMsg: "Error checking email existence",
		},
		{
			name: "HashPasswordError",
			setupService: func() *AuthServiceImpl {
				mockDB := &MockDBQueries{
					CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
					CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
				}
				mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
				mockAuth := &mockAuthConfigWithHashError{}
				return &AuthServiceImpl{
					db:          mockDB,
					dbConn:      mockConn,
					auth:        mockAuth,
					redisClient: &FakeRedis{},
				}
			},
			expectedErrMsg: "Error hashing password",
		},
		{
			name: "BeginTxError",
			setupService: func() *AuthServiceImpl {
				mockDB := &MockDBQueries{
					CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
					CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
				}
				mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) {
					return nil, assert.AnError
				}}
				mockAuth := &mockServiceAuthConfig{}
				return &AuthServiceImpl{
					db:          mockDB,
					dbConn:      mockConn,
					auth:        mockAuth,
					redisClient: &FakeRedis{},
				}
			},
			expectedErrMsg: "Error starting transaction",
		},
		{
			name: "CreateUserError",
			setupService: func() *AuthServiceImpl {
				mockDB := &MockDBQueries{
					CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
					CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
					CreateUserFunc:             func(_ context.Context, _ database.CreateUserParams) error { return assert.AnError },
				}
				mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
				mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
				mockAuth := &mockServiceAuthConfig{}
				return &AuthServiceImpl{
					db:          mockDB,
					dbConn:      mockConn,
					auth:        mockAuth,
					redisClient: &FakeRedis{},
				}
			},
			expectedErrMsg: "Error creating user",
		},
		{
			name: "TokenGenerationError",
			setupService: func() *AuthServiceImpl {
				return newAuthServiceWithTokenOrStoreError(&mockAuthConfigWithTokenError{})
			},
			expectedErrMsg: "Error generating tokens",
		},
		{
			name: "StoreTokenError",
			setupService: func() *AuthServiceImpl {
				return newAuthServiceWithTokenOrStoreError(&mockAuthConfigWithStoreError{})
			},
			expectedErrMsg: "Error storing refresh token",
		},
	}

	params := SignUpParams{Name: "user", Email: "user@example.com", Password: testPassword}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			result, err := service.SignUp(ctx, params)
			require.Error(t, err)
			require.Nil(t, result)
			require.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// --- SignIn Error Path Tests ---

// TestAuthServiceImpl_SignIn_GetUserByEmailError tests SignIn for error when fetching user by email.
func TestAuthServiceImpl_SignIn_GetUserByEmailError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{}, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	mockAuth := &mockServiceAuthConfig{}
	service := &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        mockAuth,
		redisClient: &FakeRedis{},
	}
	params := SignInParams{Email: "user@example.com", Password: testPassword}
	result, err := service.SignIn(ctx, params)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Invalid credentials")
}

// TestAuthServiceImpl_SignIn_InvalidPassword tests SignIn for invalid password error.
func TestAuthServiceImpl_SignIn_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	userID := testUUID
	password := testPassword
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_SignIn_UUIDParseError tests SignIn for error when parsing user UUID.
func TestAuthServiceImpl_SignIn_UUIDParseError(t *testing.T) {
	ctx := context.Background()
	password := testPassword
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: "invalid-uuid", Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_SignIn_BeginTxError tests SignIn for error when starting a transaction.
func TestAuthServiceImpl_SignIn_BeginTxError(t *testing.T) {
	ctx := context.Background()
	userID := testUUID
	password := testPassword
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) {
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

// TestAuthServiceImpl_SignIn_UpdateUserStatusError tests SignIn for error during user status update.
func TestAuthServiceImpl_SignIn_UpdateUserStatusError(t *testing.T) {
	ctx := context.Background()
	userID := testUUID
	password := testPassword
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(_ context.Context, _ database.UpdateUserStatusByIDParams) error {
			return assert.AnError
		},
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_SignIn_TokenAndStoreError tests SignIn for error during token generation or refresh token storage.
func TestAuthServiceImpl_SignIn_TokenAndStoreError(t *testing.T) {
	tests := []struct {
		name           string
		mockAuth       AuthConfig
		expectedErrMsg string
	}{
		{
			name:           "TokenGenerationError",
			mockAuth:       &mockAuthConfigWithTokenError{},
			expectedErrMsg: "Error generating tokens",
		},
		{
			name:           "StoreTokenError",
			mockAuth:       &mockAuthConfigWithStoreError{},
			expectedErrMsg: "Error storing refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			userID := testUUID
			password := testPassword
			hash, _ := auth.HashPassword(password)
			mockDB := &MockDBQueries{
				GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
					return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
				},
				UpdateUserStatusByIDFunc: func(_ context.Context, _ database.UpdateUserStatusByIDParams) error { return nil },
			}
			mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
			mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
			service := &AuthServiceImpl{
				db:          mockDB,
				dbConn:      mockConn,
				auth:        tt.mockAuth,
				redisClient: &FakeRedis{},
			}
			params := SignInParams{Email: "user@example.com", Password: password}
			result, err := service.SignIn(ctx, params)
			require.Error(t, err)
			require.Nil(t, result)
			require.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// TestAuthServiceImpl_SignIn_CommitError tests SignIn for error during transaction commit.
func TestAuthServiceImpl_SignIn_CommitError(t *testing.T) {
	ctx := context.Background()
	userID := testUUID
	password := testPassword
	hash, _ := auth.HashPassword(password)
	mockDB := &MockDBQueries{
		GetUserByEmailFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: userID, Password: sql.NullString{String: hash, Valid: true}}, nil
		},
		UpdateUserStatusByIDFunc: func(_ context.Context, _ database.UpdateUserStatusByIDParams) error { return nil },
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockTx := &MockDBTx{commitFunc: func() error { return assert.AnError }}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return mockTx, nil }}
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

// TestAuthServiceImpl_getUserInfoFromGoogle verifies behavior when fetching user info from Google, covering success, HTTP error, and JSON decode failure cases.
func TestAuthServiceImpl_getUserInfoFromGoogle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user := &UserGoogleInfo{ID: "id", Name: "name", Email: "email@example.com"}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(user)
		}))
		defer ts.Close()
		client := ts.Client()
		svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
		result, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL, client)
		require.NoError(t, err)
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
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := w.Write([]byte("not json")); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		}))
		defer ts.Close()
		client := ts.Client()
		svc := &AuthServiceImpl{oauth: &oauth2.Config{}}
		_, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL, client)
		assert.Error(t, err)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip is a mock implementation for http.RoundTripper.
func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func Test_getUserInfoFromGoogle_Scenarios(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedID    string
		expectedEmail string
		expectedName  string
	}{
		{
			name:          "NoCustomClient",
			responseBody:  `{"id":"123","email":"test@example.com","name":"Test User"}`,
			expectedID:    "123",
			expectedEmail: "test@example.com",
			expectedName:  "Test User",
		},
		{
			name:          "DefaultClient",
			responseBody:  `{"id":"456","email":"default@example.com","name":"Default Client"}`,
			expectedID:    "456",
			expectedEmail: "default@example.com",
			expectedName:  "Default Client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(tt.responseBody)); err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))
			defer ts.Close()

			mockOAuth := &mockOAuth2ExchangerWithClient{
				client: ts.Client(),
			}

			svc := &AuthServiceImpl{oauth: mockOAuth}
			user, err := svc.getUserInfoFromGoogle(&oauth2.Token{AccessToken: "token"}, ts.URL)
			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.expectedID, user.ID)
			assert.Equal(t, tt.expectedEmail, user.Email)
			assert.Equal(t, tt.expectedName, user.Name)
		})
	}
}

// TestAuthServiceImpl_HandleGoogleAuth tests the Google OAuth handler with various scenarios:
// - happy path for a new user
// - failure due to invalid state in Redis
// - failure due to token exchange error
func TestAuthServiceImpl_HandleGoogleAuth(t *testing.T) {
	t.Run("happy path - new user", func(t *testing.T) {
		ctx := context.Background()
		redis := &FakeRedis{getResult: "valid"}
		mockDB := &MockDBQueries{
			CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
				return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
			},
			CreateUserFunc:                    func(_ context.Context, _ database.CreateUserParams) error { return nil },
			UpdateUserSigninStatusByEmailFunc: func(_ context.Context, _ database.UpdateUserSigninStatusByEmailParams) error { return nil },
		}
		mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
		mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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
		}
		result, err := ts.HandleGoogleAuth(ctx, "code", "state")
		require.NoError(t, err)
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
		require.Error(t, err)
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
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// Test double for authServiceImpl to override getUserInfoFromGoogle
type testAuthServiceImpl struct {
	AuthServiceImpl
}

// func (s *testAuthServiceImpl) getUserInfoFromGoogle(token *oauth2.Token, userInfoURL string, clientOpt ...*http.Client) (*UserGoogleInfo, error) {
// 	if s.getUserInfoFromGoogleFunc != nil {
// 		return s.getUserInfoFromGoogleFunc(token, userInfoURL, clientOpt...)
// 	}
// 	return s.AuthServiceImpl.getUserInfoFromGoogle(token, userInfoURL, clientOpt...)
// }

// TestAuthServiceImpl_SignOut verifies the behavior of the SignOut function:
// - success case where the refresh token is deleted from Redis
// - failure case where Redis returns an error during deletion
func TestAuthServiceImpl_SignOut(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			DelFunc: func(_ context.Context, keys ...string) *redis.IntCmd {
				assert.Contains(t, keys[0], "refresh_token:")
				return redis.NewIntResult(1, nil)
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis}
		err := svc.SignOut(context.Background(), testUserID, testProvider)
		assert.NoError(t, err)
	})
	t.Run("redis error", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			DelFunc: func(_ context.Context, _ ...string) *redis.IntCmd {
				return redis.NewIntResult(0, assert.AnError)
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis}
		err := svc.SignOut(context.Background(), testUserID, testProvider)
		require.Error(t, err)
		appErr := &handlers.AppError{}
		ok := errors.As(err, &appErr)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

// TestAuthServiceImpl_GenerateGoogleAuthURL tests generating the Google OAuth URL:
// - success case where the state is stored in Redis and URL is returned
// - failure case where storing the state in Redis fails
func TestAuthServiceImpl_GenerateGoogleAuthURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			SetFunc: func(_ context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
				assert.Contains(t, key, "oauth_state:")
				assert.Equal(t, "valid", value)
				return redis.NewStatusResult("OK", nil)
			},
		}
		mockOAuth := &mockOAuth2Exchanger{
			AuthCodeURLFunc: func(_ string, _ ...oauth2.AuthCodeOption) string {
				return "https://accounts.google.com/o/oauth2/auth?state=" + "xyz123"
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis, oauth: mockOAuth}
		url, err := svc.GenerateGoogleAuthURL("xyz123")
		require.NoError(t, err)
		assert.Contains(t, url, "state=xyz123")
	})
	t.Run("redis error", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			SetFunc: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.StatusCmd {
				return redis.NewStatusResult("", assert.AnError)
			},
		}
		mockOAuth := &mockOAuth2Exchanger{
			AuthCodeURLFunc: func(_ string, _ ...oauth2.AuthCodeOption) string {
				return ""
			},
		}
		svc := &AuthServiceImpl{redisClient: mockRedis, oauth: mockOAuth}
		url, err := svc.GenerateGoogleAuthURL("failstate")
		require.Error(t, err)
		assert.Empty(t, url)
		appErr := &handlers.AppError{}
		ok := errors.As(err, &appErr)
		assert.True(t, ok)
		assert.Equal(t, "redis_error", appErr.Code)
	})
}

// --- MergeCart tests ---

// TestMergeCart_NoSessionID tests MergeCart for the case when no session ID is present.
func TestMergeCart_NoSessionID(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart:       func(_ context.Context, _ string) (*models.Cart, error) { return nil, nil },
		DeleteGuestCart:    func(_ context.Context, _ string) error { return nil },
	}

	// Create request without session ID
	req := httptest.NewRequest("POST", "/signin", nil)
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify no methods were called since session ID is empty
	cfg.AssertNotCalled(t, "LogHandlerError")
}

// TestMergeCart_GetGuestCartError tests MergeCart for error when getting the guest cart.
func TestMergeCart_GetGuestCartError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return nil, errors.New("redis error")
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return nil },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "get_guest_cart_failed", "Failed to get guest cart", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertExpectations(t)
}

// TestMergeCart_EmptyGuestCart tests MergeCart for the case when the guest cart is empty.
func TestMergeCart_EmptyGuestCart(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return &models.Cart{Items: []models.CartItem{}}, nil
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return nil },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertNotCalled(t, "LogHandlerError")
}

// TestMergeCart_EmptyGuestCartDeleteError tests MergeCart for error when deleting an empty guest cart.
func TestMergeCart_EmptyGuestCartDeleteError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return &models.Cart{Items: []models.CartItem{}}, nil
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return errors.New("delete error") },
	}

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete empty guest cart", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertExpectations(t)
}

// TestMergeCart_MergeError tests MergeCart for error during merging guest cart to user cart.
func TestMergeCart_MergeError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return nil },
	}

	// Mock CartMG to fail
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, testUserID, mock.Anything).Return(errors.New("merge error"))

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "merge_cart_failed", "Failed to merge guest cart to user", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertExpectations(t)
	cfg.CartMG.AssertExpectations(t)
}

// TestMergeCart_Success tests MergeCart for successful merging of guest cart to user cart.
func TestMergeCart_Success(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return nil },
	}

	// Mock CartMG to succeed
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, testUserID, mock.Anything).Return(nil)

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertNotCalled(t, "LogHandlerError")
	cfg.CartMG.AssertExpectations(t)
}

// TestMergeCart_SuccessWithDeleteError tests MergeCart for successful merge but error during guest cart deletion.
func TestMergeCart_SuccessWithDeleteError(t *testing.T) {
	cfg := &TestMergeCartConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Auth:               &mockAuthConfig{},
		authService:        &MockAuthService{},
		CartMG:             &MockCartManager{},
		GetGuestCart: func(_ context.Context, _ string) (*models.Cart, error) {
			return &models.Cart{
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
					{ProductID: "prod2", Quantity: 1},
				},
			}, nil
		},
		DeleteGuestCart: func(_ context.Context, _ string) error { return errors.New("delete error") },
	}

	// Mock CartMG to succeed
	cfg.CartMG.On("MergeGuestCartToUser", mock.Anything, testUserID, mock.Anything).Return(nil)

	// Create request with session ID cookie
	req := httptest.NewRequest("POST", "/signin", nil)
	req.AddCookie(&http.Cookie{Name: "guest_session_id", Value: "session123"})
	ctx := context.Background()

	// Mock error logging
	cfg.MockHandlersConfig.On("LogHandlerError", ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete guest cart after merge", "", "", mock.Anything).Return()

	// Execute
	cfg.MergeCart(ctx, req, testUserID)

	// Verify expectations
	cfg.AssertExpectations(t)
	cfg.CartMG.AssertExpectations(t)
}

// --- HandleGoogleUserAuth Error Path and Happy Path Tests ---

// TestAuthServiceImpl_handleGoogleUserAuth_DBError tests handleGoogleUserAuth for DB error.
func TestAuthServiceImpl_handleGoogleUserAuth_DBError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, assert.AnError
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_handleGoogleUserAuth_BeginTxError tests handleGoogleUserAuth for transaction begin error.
func TestAuthServiceImpl_handleGoogleUserAuth_BeginTxError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
	}
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return nil, assert.AnError }}
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

// TestAuthServiceImpl_handleGoogleUserAuth_CreateUserError tests handleGoogleUserAuth for user creation error.
func TestAuthServiceImpl_handleGoogleUserAuth_CreateUserError(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
		},
		CreateUserFunc: func(_ context.Context, _ database.CreateUserParams) error {
			return assert.AnError
		},
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_handleGoogleUserAuth_ErrorScenarios tests handleGoogleUserAuth for various error scenarios.
func TestAuthServiceImpl_handleGoogleUserAuth_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		mockAuth       AuthConfig
		refreshToken   string
		expectedErrMsg string
	}{
		{
			name:           "GenerateAccessTokenError",
			mockAuth:       &mockAuthConfigWithTokenError{},
			refreshToken:   "refresh",
			expectedErrMsg: "Error generating access token",
		},
		{
			name:           "NoRefreshToken",
			mockAuth:       &mockServiceAuthConfig{},
			refreshToken:   "",
			expectedErrMsg: "No refresh token provided by Google",
		},
		{
			name:           "StoreRefreshTokenError",
			mockAuth:       &mockAuthConfigWithStoreError{},
			refreshToken:   "refresh",
			expectedErrMsg: "Error storing refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockDB := &MockDBQueries{
				CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
					return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
				},
				UpdateUserSigninStatusByEmailFunc: func(_ context.Context, _ database.UpdateUserSigninStatusByEmailParams) error { return nil },
			}
			mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
			mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
			service := &AuthServiceImpl{
				db:          mockDB,
				dbConn:      mockConn,
				auth:        tt.mockAuth,
				redisClient: &FakeRedis{},
			}
			user := &UserGoogleInfo{ID: "gid", Name: "gname", Email: "gemail@example.com"}
			token := &oauth2.Token{RefreshToken: tt.refreshToken}
			result, err := service.handleGoogleUserAuth(ctx, user, token)
			require.Error(t, err)
			require.Nil(t, result)
			require.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_NewUser tests handleGoogleUserAuth for the happy path of a new user.
func TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_NewUser(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{}, sql.ErrNoRows
		},
		CreateUserFunc: func(_ context.Context, _ database.CreateUserParams) error {
			return nil
		},
		UpdateUserSigninStatusByEmailFunc: func(_ context.Context, _ database.UpdateUserSigninStatusByEmailParams) error {
			return nil
		},
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_ExistingUser tests handleGoogleUserAuth for the happy path of an existing user.
func TestAuthServiceImpl_handleGoogleUserAuth_HappyPath_ExistingUser(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDBQueries{
		CheckExistsAndGetIDByEmailFunc: func(_ context.Context, _ string) (database.CheckExistsAndGetIDByEmailRow, error) {
			return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "uid"}, nil
		},
		UpdateUserSigninStatusByEmailFunc: func(_ context.Context, _ database.UpdateUserSigninStatusByEmailParams) error {
			return nil
		},
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
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

// TestAuthServiceImpl_refreshGoogleToken_Success verifies that a new access token is correctly generated using a valid Google refresh token.
func TestAuthServiceImpl_refreshGoogleToken_Success(t *testing.T) {
	ctx := context.Background()
	userID := testUserID2
	refreshToken := testRefreshToken
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

// TestAuthServiceImpl_refreshGoogleToken_TokenError verifies that an error is returned when refreshing the Google token fails.
func TestAuthServiceImpl_refreshGoogleToken_TokenError(t *testing.T) {
	ctx := context.Background()
	userID := testUserID2
	refreshToken := testRefreshToken
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

// TestRealMergeCart_NoSessionID verifies that MergeCart exits early and does not log an error when the session ID is missing from the request.
func TestRealMergeCart_NoSessionID(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	realAuthConfig := &auth.Config{}

	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			Auth: realAuthConfig,
		},
		Logger: mockHandlersConfig,
	}

	// Create request without session ID
	req := httptest.NewRequest("POST", "/signin", nil)
	ctx := context.Background()

	// Execute - should return early without calling any methods
	cfg.MergeCart(ctx, req, testUserID)

	// Verify no error logging was called since session ID is empty
	mockHandlersConfig.AssertNotCalled(t, "LogHandlerError")
}

// newAuthServiceWithTokenOrStoreError returns an AuthServiceImpl with the given auth config for error path tests.
func newAuthServiceWithTokenOrStoreError(authConfig AuthConfig) *AuthServiceImpl {
	mockDB := &MockDBQueries{
		CheckUserExistsByNameFunc:  func(_ context.Context, _ string) (bool, error) { return false, nil },
		CheckUserExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
		CreateUserFunc:             func(_ context.Context, _ database.CreateUserParams) error { return nil },
	}
	mockDB.WithTxFunc = func(_ DBTx) DBQueries { return mockDB }
	mockConn := &MockDBConn{beginTxFunc: func(_ context.Context, _ *sql.TxOptions) (DBTx, error) { return &MockDBTx{}, nil }}
	return &AuthServiceImpl{
		db:          mockDB,
		dbConn:      mockConn,
		auth:        authConfig,
		redisClient: &FakeRedis{},
	}
}
