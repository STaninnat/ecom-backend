package authhandlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// contextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	HttpRequestKey ContextKey = "httpRequest"
)

// DBQueriesAdapter adapts *database.Queries to DBQueries interface
// (You may need to add more methods as needed)
type DBQueriesAdapter struct {
	*database.Queries
}

func (a *DBQueriesAdapter) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return a.Queries.CheckUserExistsByName(ctx, name)
}
func (a *DBQueriesAdapter) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.Queries.CheckUserExistsByEmail(ctx, email)
}
func (a *DBQueriesAdapter) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return a.Queries.CreateUser(ctx, params)
}
func (a *DBQueriesAdapter) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return a.Queries.GetUserByEmail(ctx, email)
}
func (a *DBQueriesAdapter) UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
	return a.Queries.UpdateUserStatusByID(ctx, params)
}
func (a *DBQueriesAdapter) WithTx(tx DBTx) DBQueries {
	return &DBQueriesAdapter{a.Queries.WithTx(tx.(*sql.Tx))}
}
func (a *DBQueriesAdapter) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	return a.Queries.CheckExistsAndGetIDByEmail(ctx, email)
}
func (a *DBQueriesAdapter) UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
	return a.Queries.UpdateUserSigninStatusByEmail(ctx, params)
}

// DBConnAdapter adapts *sql.DB to DBConn interface
type DBConnAdapter struct {
	*sql.DB
}

func (a *DBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	return tx, err
}

// AuthConfigAdapter adapts *auth.AuthConfig to AuthConfig interface
type AuthConfigAdapter struct {
	AuthConfig *auth.AuthConfig
}

// HashPassword uses the package-level function, since AuthConfig does not have a method for it.
func (a *AuthConfigAdapter) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}
func (a *AuthConfigAdapter) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return a.AuthConfig.GenerateTokens(userID, expiresAt)
}

// StoreRefreshTokenInRedis expects *http.Request, not context.Context
func (a *AuthConfigAdapter) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	if a.AuthConfig == nil {
		return errors.New("AuthConfig is nil")
	}
	r, ok := ctx.Value(HttpRequestKey).(*http.Request)
	if !ok || r == nil {
		return errors.New("StoreRefreshTokenInRedis requires *http.Request in context under 'httpRequest' key")
	}
	return a.AuthConfig.StoreRefreshTokenInRedis(r, userID, refreshToken, provider, ttl)
}
func (a *AuthConfigAdapter) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return a.AuthConfig.GenerateAccessToken(userID, expiresAt)
}
