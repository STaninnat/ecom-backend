package userhandlers

import (
	"context"
	"database/sql"
	"testing"

	"errors"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/handlers"
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
	err := &handlers.AppError{Code: "test", Message: "msg", Err: baseErr}
	assert.Equal(t, baseErr, errors.Unwrap(err))
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}

	expectedUser := database.User{
		ID:      "u1",
		Name:    "Alice",
		Email:   "alice@example.com",
		Phone:   sql.NullString{String: "123", Valid: true},
		Address: sql.NullString{String: "Addr", Valid: true},
	}

	// Mock the database query
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs("u1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "phone", "address", "password", "provider", "provider_id", "role", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.Phone.String, expectedUser.Address.String, nil, "", nil, "", time.Now(), time.Now()),
	)

	user, err := service.GetUserByID(context.Background(), "u1")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
}

func TestUserService_GetUserByID_NilDB(t *testing.T) {
	service := &userServiceImpl{db: nil, dbConn: nil}

	user, err := service.GetUserByID(context.Background(), "u1")
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "db is nil")
}

func TestUserService_GetUserByID_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}

	// Mock the database query to return an error
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs("u1").WillReturnError(errors.New("database error"))

	user, err := service.GetUserByID(context.Background(), "u1")
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "database error")
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com", Phone: "123", Address: "Addr"}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WithArgs(
		user.ID, params.Name, params.Email,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.UpdateUser(context.Background(), user, params)
	assert.NoError(t, err)
}

func TestUserService_UpdateUser_EmptyParams(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "", Email: "", Phone: "", Address: ""}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WithArgs(
		user.ID, params.Name, params.Email,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.UpdateUser(context.Background(), user, params)
	assert.NoError(t, err)
}

func TestUserService_UpdateUser_PartialParams(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"} // Phone and Address empty

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WithArgs(
		user.ID, params.Name, params.Email,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.UpdateUser(context.Background(), user, params)
	assert.NoError(t, err)
}

func TestUserService_UpdateUser_BeginTxError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Bob", Email: "bob@example.com"}

	mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error starting transaction")
}

func TestUserService_UpdateUser_UpdateUserInfoError_WithRollback(t *testing.T) {
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

func TestUserService_UpdateUser_CommitError_WithRollback(t *testing.T) {
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

// TODO: For full transaction and update tests, use sqlmock or a real *sql.DB with a test database.
