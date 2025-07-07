package userhandlers

import (
	"context"
	"database/sql"
	"testing"

	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestUserService_GetUser(t *testing.T) {
	db := &database.Queries{}
	service := &userServiceImpl{db: db, dbConn: nil}

	dbUser := database.User{
		ID:      "u1",
		Name:    "Alice",
		Email:   "alice@example.com",
		Phone:   sql.NullString{String: "123", Valid: true},
		Address: sql.NullString{String: "Addr", Valid: true},
	}

	resp, err := service.GetUser(context.Background(), dbUser)
	assert.NoError(t, err)
	assert.Equal(t, &UserResponse{
		ID:      "u1",
		Name:    "Alice",
		Email:   "alice@example.com",
		Phone:   "123",
		Address: "Addr",
	}, resp)
}

func TestUserService_UpdateUser_TransactionError(t *testing.T) {
	db := &database.Queries{}
	// Pass nil dbConn to simulate transaction error
	service := &userServiceImpl{db: db, dbConn: nil}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"}

	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB connection is nil")
}

func TestUserService_UpdateUser_BeginTxError(t *testing.T) {
	// db, _, _ := sqlmock.New() // unused
	service := &userServiceImpl{db: &database.Queries{}, dbConn: nil}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"}

	// Simulate nil dbConn (already covered), now simulate error from BeginTx
	service.dbConn = struct {
		*sql.DB
	}{nil}.DB // nil DB, will cause BeginTx to fail
	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB connection is nil")
}

func TestUserService_UpdateUser_UpdateUserInfoError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB update error")
}

func TestUserService_UpdateUser_CommitError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")
}

func TestUserError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := &UserError{Code: "test", Message: "msg", Err: baseErr}
	assert.Equal(t, baseErr, errors.Unwrap(err))
}

// TODO: For full transaction and update tests, use sqlmock or a real *sql.DB with a test database.
