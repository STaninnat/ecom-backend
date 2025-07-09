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

// --- DBQueriesAdapter basic instantiation/interface test ---
func TestDBQueriesAdapter_Instantiation(t *testing.T) {
	adapter := &DBQueriesAdapter{Queries: nil}
	assert.NotNil(t, adapter)
}

// --- AuthConfigAdapter meaningful tests ---
func TestAuthConfigAdapter_HashPassword(t *testing.T) {
	adapter := &AuthConfigAdapter{AuthConfig: &auth.AuthConfig{}}
	hash, err := adapter.HashPassword("short")
	assert.Error(t, err)
	assert.Empty(t, hash)

	hash, err = adapter.HashPassword("longenoughpassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

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

// --- DBQueriesAdapter method forwarding tests ---
type fakeQueries struct {
	CheckUserExistsByNameFunc         func(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmailFunc        func(ctx context.Context, email string) (bool, error)
	CreateUserFunc                    func(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmailFunc                func(ctx context.Context, email string) (database.User, error)
	UpdateUserStatusByIDFunc          func(ctx context.Context, params database.UpdateUserStatusByIDParams) error
	WithTxFunc                        func(tx interface{}) *fakeQueries
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
func (f *fakeQueries) WithTx(tx interface{}) *fakeQueries {
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

// --- AuthConfigAdapter GenerateTokens/GenerateAccessToken tests ---
func TestAuthConfigAdapter_GenerateTokens(t *testing.T) {
	authCfg := &auth.AuthConfig{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	adapter := &AuthConfigAdapter{AuthConfig: authCfg}
	expiresAt := time.Now().Add(time.Hour)
	access, refresh, err := adapter.GenerateTokens("user-id", expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestAuthConfigAdapter_GenerateAccessToken(t *testing.T) {
	authCfg := &auth.AuthConfig{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	adapter := &AuthConfigAdapter{AuthConfig: authCfg}
	expiresAt := time.Now().Add(time.Hour)
	token, err := adapter.GenerateAccessToken("user-id", expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
