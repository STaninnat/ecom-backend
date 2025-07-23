// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
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

// auth_adapters.go: Adapter layer to bridge underlying database, authentication, and HTTP request handling
// providing unified interfaces for DB queries, transactions, and auth token management.

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// HTTPRequestKey is the context key for storing *http.Request in context.
const (
	HTTPRequestKey ContextKey = "httpRequest"
)

// DBQueriesAdapter adapts *database.Queries to the DBQueries interface.
type DBQueriesAdapter struct {
	*database.Queries
}

// CheckUserExistsByName checks if a user exists by name.
func (a *DBQueriesAdapter) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return a.Queries.CheckUserExistsByName(ctx, name)
}

// CheckUserExistsByEmail checks if a user exists by email.
func (a *DBQueriesAdapter) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.Queries.CheckUserExistsByEmail(ctx, email)
}

// CreateUser creates a new user in the database.
func (a *DBQueriesAdapter) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return a.Queries.CreateUser(ctx, params)
}

// GetUserByEmail retrieves a user by email.
func (a *DBQueriesAdapter) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return a.Queries.GetUserByEmail(ctx, email)
}

// UpdateUserStatusByID updates a user's status by ID.
func (a *DBQueriesAdapter) UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error {
	return a.Queries.UpdateUserStatusByID(ctx, params)
}

// WithTx returns a new DBQueries using the provided transaction.
func (a *DBQueriesAdapter) WithTx(tx DBTx) DBQueries {
	return &DBQueriesAdapter{a.Queries.WithTx(tx.(*sql.Tx))}
}

// CheckExistsAndGetIDByEmail checks if a user exists by email and returns the ID.
func (a *DBQueriesAdapter) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error) {
	return a.Queries.CheckExistsAndGetIDByEmail(ctx, email)
}

// UpdateUserSigninStatusByEmail updates a user's signin status by email.
func (a *DBQueriesAdapter) UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error {
	return a.Queries.UpdateUserSigninStatusByEmail(ctx, params)
}

// DBConnAdapter adapts *sql.DB to the DBConn interface.
type DBConnAdapter struct {
	*sql.DB
}

// BeginTx begins a new transaction.
func (a *DBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	return tx, err
}

// AuthConfigAdapter adapts *auth.Config to the AuthConfig interface.
type AuthConfigAdapter struct {
	AuthConfig *auth.Config
}

// HashPassword uses the package-level function, since AuthConfig does not have a method for it.
func (a *AuthConfigAdapter) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

// GenerateTokens generates access and refresh tokens for a user.
func (a *AuthConfigAdapter) GenerateTokens(userID string, expiresAt time.Time) (string, string, error) {
	return a.AuthConfig.GenerateTokens(userID, expiresAt)
}

// StoreRefreshTokenInRedis expects *http.Request, not context.Context
func (a *AuthConfigAdapter) StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error {
	if a.AuthConfig == nil {
		return errors.New("AuthConfig is nil")
	}
	r, ok := ctx.Value(HTTPRequestKey).(*http.Request)
	if !ok || r == nil {
		return errors.New("StoreRefreshTokenInRedis requires *http.Request in context under 'httpRequest' key")
	}
	return a.AuthConfig.StoreRefreshTokenInRedis(r, userID, refreshToken, provider, ttl)
}

// GenerateAccessToken generates an access token for a user.
func (a *AuthConfigAdapter) GenerateAccessToken(userID string, expiresAt time.Time) (string, error) {
	return a.AuthConfig.GenerateAccessToken(userID, expiresAt)
}
