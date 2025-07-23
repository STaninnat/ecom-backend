// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// user_service.go: Implements user business logic including retrieval, updates, role promotion, and transaction management.

// UserService defines the business logic interface for user operations.
// Provides methods for user retrieval, updates, and role management with proper error handling.
// Add more methods as needed for user-related features (e.g., GetProfile, UpdateProfile, etc.).
type UserService interface {
	GetUser(ctx context.Context, user database.User) (*UserResponse, error)
	UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error
	GetUserByID(ctx context.Context, id string) (database.User, error)
	PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error
}

// UpdateUserParams represents parameters for updating user information.
// Contains all fields that can be updated for a user profile.
type UpdateUserParams struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

// UserResponse represents the user data returned to the client.
// Structured response format for user information with optional fields.
type UserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

// userServiceImpl implements UserService.
// Provides business logic for user operations with database transaction management.
type userServiceImpl struct {
	db     *database.Queries
	dbConn *sql.DB
}

// NewUserService creates a new UserService instance.
// Factory function for creating user service instances with database dependencies.
// Parameters:
//   - db: *database.Queries for database operations
//   - dbConn: *sql.DB for transaction management
//
// Returns:
//   - UserService: configured user service instance
func NewUserService(db *database.Queries, dbConn *sql.DB) UserService {
	return &userServiceImpl{
		db:     db,
		dbConn: dbConn,
	}
}

// GetUser returns the user info as a response struct.
// Maps database user model to client-friendly response format.
// Parameters:
//   - ctx: context.Context for the operation
//   - user: database.User to convert to response format
//
// Returns:
//   - *UserResponse: formatted user data for client consumption
//   - error: nil on success, error on failure
func (s *userServiceImpl) GetUser(_ context.Context, user database.User) (*UserResponse, error) {
	return &UserResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Phone:   user.Phone.String,
		Address: user.Address.String,
	}, nil
}

// UpdateUser updates the user's information in the database.
// Uses database transactions to ensure data consistency and proper error handling.
// Parameters:
//   - ctx: context.Context for the operation
//   - user: database.User to update
//   - params: UpdateUserParams containing the update data
//
// Returns:
//   - error: nil on success, AppError with appropriate code on failure
func (s *userServiceImpl) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)

	err = queries.UpdateUserInfo(ctx, database.UpdateUserInfoParams{
		ID:        user.ID,
		Name:      params.Name,
		Email:     params.Email,
		Phone:     utils.ToNullString(params.Phone),
		Address:   utils.ToNullString(params.Address),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return &handlers.AppError{Code: "update_failed", Message: "DB update error", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// PromoteUserToAdmin promotes a user to admin, only if the acting user is admin.
// Validates admin privileges, checks target user existence, and updates role in database transaction.
// Parameters:
//   - ctx: context.Context for the operation
//   - adminUser: database.User representing the admin performing the action
//   - targetUserID: string identifier of the user to promote
//
// Returns:
//   - error: nil on success, AppError with appropriate code on failure
func (s *userServiceImpl) PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error {
	if adminUser.Role != "admin" {
		return &handlers.AppError{Code: "unauthorized_user", Message: "Admin privileges required"}
	}
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)
	targetUser, err := queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return &handlers.AppError{Code: "user_not_found", Message: "Target user not found", Err: err}
	}
	if targetUser.Role == "admin" {
		return &handlers.AppError{Code: "already_admin", Message: "User is already admin"}
	}

	err = queries.UpdateUserRole(ctx, database.UpdateUserRoleParams{
		Role: "admin",
		ID:   targetUserID,
	})
	if err != nil {
		return &handlers.AppError{Code: "update_error", Message: "Failed to update user role", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}
	return nil
}

// UserError is an alias for handlers.AppError for consistency in user service.
type UserError = handlers.AppError

// GetUserByID retrieves a user by their ID from the database.
// Delegates to the underlying database queries with nil check.
// Parameters:
//   - ctx: context.Context for the operation
//   - id: string identifier of the user to retrieve
//
// Returns:
//   - database.User: the found user
//   - error: nil on success, error on failure
func (s *userServiceImpl) GetUserByID(ctx context.Context, id string) (database.User, error) {
	if s.db == nil {
		return database.User{}, errors.New("db is nil")
	}
	return s.db.GetUserByID(ctx, id)
}
