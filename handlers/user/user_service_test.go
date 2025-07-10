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

// TestUserService_GetUser tests that GetUser correctly converts a database User
// to a UserResponse with proper field mapping
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

// TestUserService_UpdateUser_TransactionError tests that UpdateUser returns an error
// when the database connection is nil
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

// TestUserService_UpdateUser_UpdateUserInfoError tests that UpdateUser returns an error
// when the database update operation fails
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

// TestUserService_UpdateUser_CommitError tests that UpdateUser returns an error
// when the transaction commit fails
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

// TestUserError_Unwrap tests that AppError correctly unwraps to the underlying error
func TestUserError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := &handlers.AppError{Code: "test", Message: "msg", Err: baseErr}
	assert.Equal(t, baseErr, errors.Unwrap(err))
}

// TestUserService_GetUserByID_Success tests that GetUserByID successfully retrieves
// a user from the database
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

// TestUserService_GetUserByID_NilDB tests that GetUserByID returns an error
// when the database is nil
func TestUserService_GetUserByID_NilDB(t *testing.T) {
	service := &userServiceImpl{db: nil, dbConn: nil}

	user, err := service.GetUserByID(context.Background(), "u1")
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "db is nil")
}

// TestUserService_GetUserByID_DBError tests that GetUserByID returns an error
// when the database query fails
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

// TestUserService_UpdateUser_Success tests that UpdateUser successfully updates
// a user with valid parameters
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

// TestUserService_UpdateUser_EmptyParams tests that UpdateUser successfully handles
// empty parameters without errors
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

// TestUserService_UpdateUser_PartialParams tests that UpdateUser successfully handles
// partial parameters (some fields empty)
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

// TestUserService_UpdateUser_BeginTxError tests that UpdateUser returns an error
// when beginning the transaction fails
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

// TestUserService_UpdateUser_UpdateUserInfoError_WithRollback tests that UpdateUser
// properly rolls back the transaction when the update fails
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

// TestUserService_UpdateUser_CommitError_WithRollback tests that UpdateUser
// properly handles commit errors
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

// TestUserService_PromoteUserToAdmin_Success tests that PromoteUserToAdmin successfully
// promotes a user to admin role
func TestUserService_PromoteUserToAdmin_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	targetUserID := "user2"
	targetUser := database.User{ID: targetUserID, Role: "user"}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(targetUserID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "phone", "address", "password", "provider", "provider_id", "role", "created_at", "updated_at"}).
			AddRow(targetUser.ID, "Target", "target@example.com", "", "", "", "local", "", targetUser.Role, time.Now(), time.Now()),
	)
	mock.ExpectExec("UPDATE users").WithArgs(targetUserID, "admin").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.PromoteUserToAdmin(context.Background(), adminUser, targetUserID)
	assert.NoError(t, err)
}

// TestUserService_PromoteUserToAdmin_Unauthorized tests that PromoteUserToAdmin returns
// an error when the requesting user is not an admin
func TestUserService_PromoteUserToAdmin_Unauthorized(t *testing.T) {
	service := &userServiceImpl{}
	nonAdmin := database.User{ID: "u1", Role: "user"}
	err := service.PromoteUserToAdmin(context.Background(), nonAdmin, "target")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Admin privileges required")
}

// TestUserService_PromoteUserToAdmin_AlreadyAdmin tests that PromoteUserToAdmin returns
// an error when the target user is already an admin
func TestUserService_PromoteUserToAdmin_AlreadyAdmin(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	targetUserID := "user2"
	targetUser := database.User{ID: targetUserID, Role: "admin"}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(targetUserID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "phone", "address", "password", "provider", "provider_id", "role", "created_at", "updated_at"}).
			AddRow(targetUser.ID, "Target", "target@example.com", "", "", "", "local", "", targetUser.Role, time.Now(), time.Now()),
	)
	mock.ExpectRollback()

	err := service.PromoteUserToAdmin(context.Background(), adminUser, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already admin")
}

// TestUserService_PromoteUserToAdmin_UserNotFound tests that PromoteUserToAdmin returns
// an error when the target user is not found
func TestUserService_PromoteUserToAdmin_UserNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	targetUserID := "user2"

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(targetUserID).WillReturnError(errors.New("not found"))
	mock.ExpectRollback()

	err := service.PromoteUserToAdmin(context.Background(), adminUser, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Target user not found")
}

// TestUserService_PromoteUserToAdmin_TransactionError tests that PromoteUserToAdmin returns
// an error when the database connection is nil
func TestUserService_PromoteUserToAdmin_TransactionError(t *testing.T) {
	service := &userServiceImpl{db: &database.Queries{}, dbConn: nil}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	err := service.PromoteUserToAdmin(context.Background(), adminUser, "target")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB connection is nil")
}

// TestUserService_PromoteUserToAdmin_UpdateError tests that PromoteUserToAdmin returns
// an error when the role update fails
func TestUserService_PromoteUserToAdmin_UpdateError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	targetUserID := "user2"

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(targetUserID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "phone", "address", "password", "provider", "provider_id", "role", "created_at", "updated_at"}).
			AddRow(targetUserID, "Target", "target@example.com", "", "", "", "local", "", "user", time.Now(), time.Now()),
	)
	mock.ExpectExec("UPDATE users").WithArgs(targetUserID, "admin").WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	err := service.PromoteUserToAdmin(context.Background(), adminUser, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update user role")
}

// TestUserService_PromoteUserToAdmin_CommitError tests that PromoteUserToAdmin returns
// an error when the transaction commit fails
func TestUserService_PromoteUserToAdmin_CommitError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}
	adminUser := database.User{ID: "admin1", Role: "admin"}
	targetUserID := "user2"

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(targetUserID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "phone", "address", "password", "provider", "provider_id", "role", "created_at", "updated_at"}).
			AddRow(targetUserID, "Target", "target@example.com", "", "", "", "local", "", "user", time.Now(), time.Now()),
	)
	mock.ExpectExec("UPDATE users").WithArgs(targetUserID, "admin").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := service.PromoteUserToAdmin(context.Background(), adminUser, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")
}

// TestUserService_GetUserByID_UserNotFound tests that GetUserByID returns an error
// when the user is not found in the database
func TestUserService_GetUserByID_UserNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)
	service := &userServiceImpl{db: queries, dbConn: db}

	mock.ExpectQuery("SELECT (.+) FROM users").WithArgs("u404").WillReturnError(sql.ErrNoRows)

	user, err := service.GetUserByID(context.Background(), "u404")
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "no rows")
}

// TestNewUserService_ReturnsNonNil tests that NewUserService returns a non-nil service
func TestNewUserService_ReturnsNonNil(t *testing.T) {
	db := &database.Queries{}
	dbConn := new(sql.DB)
	service := NewUserService(db, dbConn)
	assert.NotNil(t, service)
}

// TestUserService_GetUser_EmptyFields tests that GetUser correctly handles
// users with empty/null fields
func TestUserService_GetUser_EmptyFields(t *testing.T) {
	service := &userServiceImpl{}
	dbUser := database.User{ID: "u2"} // All other fields zero values
	resp, err := service.GetUser(context.Background(), dbUser)
	assert.NoError(t, err)
	assert.Equal(t, &UserResponse{ID: "u2", Name: "", Email: "", Phone: "", Address: ""}, resp)
}

// TestUserService_UpdateUser_NilDBAndDBConn tests that UpdateUser returns an error
// when both database and database connection are nil
func TestUserService_UpdateUser_NilDBAndDBConn(t *testing.T) {
	service := &userServiceImpl{db: nil, dbConn: nil}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "Test", Email: "test@example.com"}
	err := service.UpdateUser(context.Background(), user, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB connection is nil")
}
