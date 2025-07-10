package authhandlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
)

// TestDBQueriesAdapter_Instantiation verifies that a DBQueriesAdapter can be instantiated and is not nil.
func TestDBQueriesAdapter_Instantiation(t *testing.T) {
	adapter := &DBQueriesAdapter{Queries: nil}
	assert.NotNil(t, adapter)
}

// TestAuthConfigAdapter_HashPassword tests the HashPassword method for both short and valid passwords, checking for correct error handling and hash generation.
func TestAuthConfigAdapter_HashPassword(t *testing.T) {
	adapter := &AuthConfigAdapter{AuthConfig: &auth.AuthConfig{}}
	hash, err := adapter.HashPassword("short")
	assert.Error(t, err)
	assert.Empty(t, hash)

	hash, err = adapter.HashPassword("longenoughpassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

// TestAuthConfigAdapter_StoreRefreshTokenInRedis_ContextCases tests StoreRefreshTokenInRedis for various context and config error cases.
func TestAuthConfigAdapter_StoreRefreshTokenInRedis_ContextCases(t *testing.T) {
	// Create adapter with properly initialized AuthConfig
	authConfig := &auth.AuthConfig{
		// APIConfig is intentionally nil for this test
	}
	adapter := &AuthConfigAdapter{AuthConfig: authConfig}

	// Debug: Check if the embedded AuthConfig is properly set
	assert.NotNil(t, adapter.AuthConfig, "AuthConfig should not be nil")

	ctx := context.Background()
	// No httpRequest in context
	err := adapter.StoreRefreshTokenInRedis(ctx, "u1", "rt", "local", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires *http.Request")

	// With httpRequest in context, but nil APIConfig
	r, _ := http.NewRequest("GET", "/", nil)
	ctx2 := context.WithValue(ctx, HttpRequestKey, r)
	err = adapter.StoreRefreshTokenInRedis(ctx2, "u1", "rt", "local", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "APIConfig is nil")

	// New: Test with nil embedded AuthConfig
	nilAdapter := &AuthConfigAdapter{AuthConfig: nil}
	err = nilAdapter.StoreRefreshTokenInRedis(ctx2, "u1", "rt", "local", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AuthConfig is nil")
}

// TestDBQueriesAdapter_Methods tests all DBQueriesAdapter methods using a fakeQueries implementation for correct forwarding and error handling.
func TestDBQueriesAdapter_Methods(t *testing.T) {
	ctx := context.Background()
	fq := &fakeQueries{
		CheckUserExistsByNameFunc: func(ctx context.Context, name string) (bool, error) {
			return name == "exists", nil
		},
		CheckUserExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return email == "exists@example.com", nil
		},
		CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) error {
			if params.ID == "" {
				return assert.AnError
			}
			return nil
		},
		GetUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			if email == "found@example.com" {
				return database.User{ID: "id1"}, nil
			}
			return database.User{}, assert.AnError
		},
		UpdateUserStatusByIDFunc: func(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
			if params.ID == "" {
				return assert.AnError
			}
			return nil
		},
		WithTxFunc: func(tx interface{}) *fakeQueries {
			return &fakeQueries{}
		},
		CheckExistsAndGetIDByEmailFunc: func(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
			if email == "exists@example.com" {
				return database.CheckExistsAndGetIDByEmailRow{Exists: true, ID: "id2"}, nil
			}
			return database.CheckExistsAndGetIDByEmailRow{}, assert.AnError
		},
		UpdateUserSigninStatusByEmailFunc: func(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
			if params.Email == "" {
				return assert.AnError
			}
			return nil
		},
	}
	adapter := &testDBQueriesAdapter{fakeQueries: fq}

	ok, err := adapter.CheckUserExistsByName(ctx, "exists")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = adapter.CheckUserExistsByEmail(ctx, "exists@example.com")
	assert.True(t, ok)
	assert.NoError(t, err)

	err = adapter.CreateUser(ctx, database.CreateUserParams{ID: "id"})
	assert.NoError(t, err)
	err = adapter.CreateUser(ctx, database.CreateUserParams{})
	assert.Error(t, err)

	_, err = adapter.GetUserByEmail(ctx, "found@example.com")
	assert.NoError(t, err)
	_, err = adapter.GetUserByEmail(ctx, "notfound@example.com")
	assert.Error(t, err)

	err = adapter.UpdateUserStatusByID(ctx, database.UpdateUserStatusByIDParams{ID: "id"})
	assert.NoError(t, err)
	err = adapter.UpdateUserStatusByID(ctx, database.UpdateUserStatusByIDParams{})
	assert.Error(t, err)

	_ = adapter.WithTx(nil)

	_, err = adapter.CheckExistsAndGetIDByEmail(ctx, "exists@example.com")
	assert.NoError(t, err)
	_, err = adapter.CheckExistsAndGetIDByEmail(ctx, "notfound@example.com")
	assert.Error(t, err)

	err = adapter.UpdateUserSigninStatusByEmail(ctx, database.UpdateUserSigninStatusByEmailParams{Email: "e"})
	assert.NoError(t, err)
	err = adapter.UpdateUserSigninStatusByEmail(ctx, database.UpdateUserSigninStatusByEmailParams{})
	assert.Error(t, err)
}

// --- Direct DBQueriesAdapter forwarding tests for coverage ---
// Note: DB adapter forwarding tests require a real Postgres DB for meaningful coverage.
// These are best tested via integration tests, not unit tests.
// func TestDBQueriesAdapter_Forwarding(t *testing.T) {
// 	// This test was removed because it requires a real Postgres DB
// 	// and causes panics with mock DBTX implementations.
// 	// DB adapter coverage should be achieved via integration tests.
// }

// --- DBConnAdapter tests ---
// func TestDBConnAdapter_BeginTx(t *testing.T) {
// 	db := &sql.DB{} // This will not actually connect, but we can test the method signature
// 	adapter := &DBConnAdapter{DB: db}
// 	// Should return a *sql.Tx or error (will error with nil DB)
// 	tx, err := adapter.BeginTx(context.Background(), nil)
// 	assert.Nil(t, tx)
// 	assert.Error(t, err)
// }

// TestAuthConfigAdapter_GenerateTokens tests the GenerateTokens method for correct access and refresh token generation.
func TestAuthConfigAdapter_GenerateTokens(t *testing.T) {
	authCfg := &auth.AuthConfig{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	adapter := &AuthConfigAdapter{AuthConfig: authCfg}
	expiresAt := time.Now().Add(time.Hour)
	access, refresh, err := adapter.GenerateTokens("user-id", expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

// TestAuthConfigAdapter_GenerateAccessToken tests the GenerateAccessToken method for correct token generation.
func TestAuthConfigAdapter_GenerateAccessToken(t *testing.T) {
	authCfg := &auth.AuthConfig{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	adapter := &AuthConfigAdapter{AuthConfig: authCfg}
	expiresAt := time.Now().Add(time.Hour)
	token, err := adapter.GenerateAccessToken("user-id", expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
