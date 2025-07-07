package userhandlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// UserService defines the business logic interface for user operations
// Add more methods as needed for user-related features
// (e.g., GetProfile, UpdateProfile, etc.)
type UserService interface {
	GetUser(ctx context.Context, user database.User) (*UserResponse, error)
	UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error
}

// UpdateUserParams represents parameters for updating user info
type UpdateUserParams struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

// UserResponse represents the user data returned to the client
type UserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	db     *database.Queries
	dbConn *sql.DB
}

// NewUserService creates a new UserService instance
func NewUserService(db *database.Queries, dbConn *sql.DB) UserService {
	return &userServiceImpl{
		db:     db,
		dbConn: dbConn,
	}
}

// GetUser returns the user info as a response struct
func (s *userServiceImpl) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	return &UserResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Phone:   user.Phone.String,
		Address: user.Address.String,
	}, nil
}

// UpdateUser updates the user's information in the database
func (s *userServiceImpl) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	if s.dbConn == nil {
		return &UserError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &UserError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

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
		return &UserError{Code: "update_failed", Message: "DB update error", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &UserError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// UserError is a custom error type for user operations
// It provides a code and message for error handling
// Similar to AuthError in auth_service.go
type UserError struct {
	Code    string
	Message string
	Err     error
}

func (e *UserError) Error() string {
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Err
}
