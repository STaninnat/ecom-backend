// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// category_service_test.go: Tests for category DB query interfaces, adapters, and business logic service with transaction handling.

// TestNewCategoryService tests the NewCategoryService constructor function.
// It verifies that the service can be created with both nil and valid parameters
// without causing panics or errors. This ensures the constructor is robust
// and handles edge cases gracefully.
func TestNewCategoryService(t *testing.T) {
	// Test with nil parameters
	service := NewCategoryService(nil, nil)
	assert.NotNil(t, service)

	// Test with valid parameters
	mockDB := &database.Queries{}
	mockDBConn := &sql.DB{}
	service = NewCategoryService(mockDB, mockDBConn)
	assert.NotNil(t, service)
}

// TestCategoryServiceImpl_CreateCategory tests the CreateCategory method of the category service implementation.
// It covers various scenarios including successful creation, validation errors (empty name, name too long,
// description too long), and database errors. Each test case uses mocked database interactions
// to isolate the service logic from actual database operations.
func TestCategoryServiceImpl_CreateCategory(t *testing.T) {
	tests := []struct {
		name          string
		params        CategoryRequest
		setupMocks    func(*MockCategoryDBQueries, *MockCategoryDBConn, *MockCategoryDBTx)
		expectedError bool
		errorCode     string
	}{
		{
			name: "successful creation",
			params: CategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
			},
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("CreateCategory", mock.Anything, mock.Anything).Return(nil)
				mockTx.On("Commit").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "empty name",
			params: CategoryRequest{
				Name: "",
			},
			setupMocks: func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {
				// No mocks needed as validation fails before DB calls
			},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name: "name too long",
			params: CategoryRequest{
				Name: string(make([]byte, 101)), // 101 characters
			},
			setupMocks: func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {
				// No mocks needed as validation fails before DB calls
			},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name: "description too long",
			params: CategoryRequest{
				Name:        "Test Category",
				Description: string(make([]byte, 501)), // 501 characters
			},
			setupMocks: func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {
				// No mocks needed as validation fails before DB calls
			},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name: "database error",
			params: CategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
			},
			setupMocks:    setupDatabaseErrorForCreate,
			expectedError: true,
			errorCode:     "create_category_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCategoryServiceTest(
				t,
				func(s *categoryServiceImpl, ctx context.Context, p CategoryRequest) (string, error) {
					return s.CreateCategory(ctx, p)
				},
				tt.params,
				tt.setupMocks,
				tt.expectedError,
				tt.errorCode,
			)
		})
	}
}

// TestCategoryServiceImpl_UpdateCategory tests the UpdateCategory method of the category service implementation.
// It covers successful updates, validation errors (missing ID, missing name), and database errors.
// The test uses mocked database transactions to verify that the service properly handles
// transaction lifecycle and error conditions.
func TestCategoryServiceImpl_UpdateCategory(t *testing.T) {
	tests := []struct {
		name          string
		params        CategoryRequest
		setupMocks    func(*MockCategoryDBQueries, *MockCategoryDBConn, *MockCategoryDBTx)
		expectedError bool
		errorCode     string
	}{
		{
			name: "successful update",
			params: CategoryRequest{
				ID:          "test-id",
				Name:        "Updated Category",
				Description: "Updated Description",
			},
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("UpdateCategories", mock.Anything, mock.Anything).Return(nil)
				mockTx.On("Commit").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "missing ID",
			params: CategoryRequest{
				Name: "Test Category",
			},
			setupMocks: func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {
				// No mocks needed as validation fails before DB calls
			},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name: "missing name",
			params: CategoryRequest{
				ID: "test-id",
			},
			setupMocks: func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {
				// No mocks needed as validation fails before DB calls
			},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name: "database error",
			params: CategoryRequest{
				ID:   "test-id",
				Name: "Test Category",
			},
			setupMocks:    setupDatabaseErrorForUpdate,
			expectedError: true,
			errorCode:     "update_category_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCategoryServiceTest(
				t,
				func(s *categoryServiceImpl, ctx context.Context, p CategoryRequest) (string, error) {
					err := s.UpdateCategory(ctx, p)
					return "", err
				},
				tt.params,
				tt.setupMocks,
				tt.expectedError,
				tt.errorCode,
			)
		})
	}
}

// TestCategoryServiceImpl_DeleteCategory tests the DeleteCategory method of the category service implementation.
// It covers successful deletion, validation errors (empty category ID), and database errors.
// The test verifies that the service properly handles transaction management and error conditions
// when deleting categories from the database.
func TestCategoryServiceImpl_DeleteCategory(t *testing.T) {
	tests := []struct {
		name          string
		categoryID    string
		setupMocks    func(*MockCategoryDBQueries, *MockCategoryDBConn, *MockCategoryDBTx)
		expectedError bool
		errorCode     string
	}{
		{
			name:       "successful deletion",
			categoryID: "test-id",
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("DeleteCategory", mock.Anything, "test-id").Return(nil)
				mockTx.On("Commit").Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "empty category ID",
			categoryID:    "",
			setupMocks:    func(_ *MockCategoryDBQueries, _ *MockCategoryDBConn, _ *MockCategoryDBTx) {},
			expectedError: true,
			errorCode:     "invalid_request",
		},
		{
			name:       "database error",
			categoryID: "test-id",
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("DeleteCategory", mock.Anything, "test-id").Return(errors.New("database error"))
			},
			expectedError: true,
			errorCode:     "delete_category_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockCategoryDBQueries{}
			mockConn := &MockCategoryDBConn{}
			mockTx := &MockCategoryDBTx{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockConn, mockTx)
			}

			service := &categoryServiceImpl{
				db:     mockDB,
				dbConn: mockConn,
			}

			err := service.DeleteCategory(context.Background(), tt.categoryID)

			if tt.expectedError {
				require.Error(t, err)
				appErr := &handlers.AppError{}
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				require.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockConn.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

// TestCategoryServiceImpl_GetAllCategories tests the GetAllCategories method of the category service implementation.
// It covers successful retrieval of all categories and database errors. Unlike other operations,
// this method doesn't require transactions since it's a read-only operation. The test verifies
// that the service properly handles the retrieval of category data and error conditions.
func TestCategoryServiceImpl_GetAllCategories(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockCategoryDBQueries)
		expectedResult []database.Category
		expectedError  bool
		errorCode      string
	}{
		{
			name: "successful retrieval",
			setupMocks: func(mockDB *MockCategoryDBQueries) {
				expectedCategories := []database.Category{
					{
						ID:          "cat1",
						Name:        "Category 1",
						Description: utils.ToNullString("Description 1"),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          "cat2",
						Name:        "Category 2",
						Description: utils.ToNullString("Description 2"),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				mockDB.On("GetAllCategories", mock.Anything).Return(expectedCategories, nil)
			},
			expectedResult: []database.Category{
				{
					ID:          "cat1",
					Name:        "Category 1",
					Description: utils.ToNullString("Description 1"),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "cat2",
					Name:        "Category 2",
					Description: utils.ToNullString("Description 2"),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			expectedError: false,
		},
		{
			name: "database error",
			setupMocks: func(mockDB *MockCategoryDBQueries) {
				mockDB.On("GetAllCategories", mock.Anything).Return([]database.Category{}, errors.New("database error"))
			},
			expectedResult: []database.Category{},
			expectedError:  true,
			errorCode:      "database_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockCategoryDBQueries{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB)
			}

			service := &categoryServiceImpl{
				db: mockDB,
			}

			result, err := service.GetAllCategories(context.Background())

			if tt.expectedError {
				require.Error(t, err)
				appErr := &handlers.AppError{}
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, result, len(tt.expectedResult))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

// TestCategoryServiceImpl_CreateCategory_NilDBConn tests the CreateCategory method when the database connection is nil.
// This edge case test ensures that the service properly handles the scenario where the database connection
// hasn't been initialized, returning an appropriate error with the correct error code and message.
func TestCategoryServiceImpl_CreateCategory_NilDBConn(t *testing.T) {
	service := &categoryServiceImpl{
		db:     &CategoryDBQueriesAdapter{},
		dbConn: nil,
	}

	_, err := service.CreateCategory(context.Background(), CategoryRequest{
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "DB connection is nil")
	}
}

// TestCategoryServiceImpl_UpdateCategory_NilDBConn tests the UpdateCategory method when the database connection is nil.
// This edge case test ensures that the service properly handles the scenario where the database connection
// hasn't been initialized, returning an appropriate error with the correct error code and message.
func TestCategoryServiceImpl_UpdateCategory_NilDBConn(t *testing.T) {
	service := &categoryServiceImpl{
		db:     &CategoryDBQueriesAdapter{},
		dbConn: nil,
	}

	err := service.UpdateCategory(context.Background(), CategoryRequest{
		ID:          "test-id",
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "DB connection is nil")
	}
}

// TestCategoryServiceImpl_DeleteCategory_NilDBConn tests the DeleteCategory method when the database connection is nil.
// This edge case test ensures that the service properly handles the scenario where the database connection
// hasn't been initialized, returning an appropriate error with the correct error code and message.
func TestCategoryServiceImpl_DeleteCategory_NilDBConn(t *testing.T) {
	service := &categoryServiceImpl{
		db:     &CategoryDBQueriesAdapter{},
		dbConn: nil,
	}

	err := service.DeleteCategory(context.Background(), "test-id")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "DB connection is nil")
	}
}

// TestCategoryServiceImpl_GetAllCategories_NilDB tests the GetAllCategories method when the database is nil.
// This edge case test ensures that the service properly handles the scenario where the database
// hasn't been initialized, returning an appropriate error with the correct error code and message.
func TestCategoryServiceImpl_GetAllCategories_NilDB(t *testing.T) {
	service := &categoryServiceImpl{
		db:     nil,
		dbConn: &CategoryDBConnAdapter{},
	}

	_, err := service.GetAllCategories(context.Background())

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "database_error", appErr.Code)
		assert.Contains(t, appErr.Message, "DB is nil")
	}
}

// TestCategoryServiceImpl_CreateCategory_TransactionError tests the CreateCategory method when the database transaction
// fails to start. This test verifies that the service properly handles transaction initialization errors
// and returns the appropriate error with the correct error code and message.
func TestCategoryServiceImpl_CreateCategory_TransactionError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}

	// Mock BeginTx to return an error
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return((*MockCategoryDBTx)(nil), fmt.Errorf("transaction failed"))

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	_, err := service.CreateCategory(context.Background(), CategoryRequest{
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error starting transaction")
	}
	mockDBConn.AssertExpectations(t)
}

// TestCategoryServiceImpl_CreateCategory_CommitError tests the CreateCategory method when the database transaction
// commit operation fails. This test verifies that the service properly handles commit errors and ensures
// that rollback is called when commit fails. It also checks that the appropriate error is returned.
func TestCategoryServiceImpl_CreateCategory_CommitError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}
	mockTx := &MockCategoryDBTx{}

	// Mock successful BeginTx
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock successful CreateCategory
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("CreateCategory", mock.Anything, mock.Anything).Return(nil)
	// Mock Commit to return an error
	mockTx.On("Commit").Return(fmt.Errorf("commit failed"))
	mockTx.On("Rollback").Return(nil)

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	_, err := service.CreateCategory(context.Background(), CategoryRequest{
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "commit_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error committing transaction")
	}
	mockDBConn.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// TestCategoryServiceImpl_UpdateCategory_TransactionError tests the UpdateCategory method when the database transaction
// fails to start. This test verifies that the service properly handles transaction initialization errors
// and returns the appropriate error with the correct error code and message.
func TestCategoryServiceImpl_UpdateCategory_TransactionError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}

	// Mock BeginTx to return an error
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return((*MockCategoryDBTx)(nil), fmt.Errorf("transaction failed"))

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	err := service.UpdateCategory(context.Background(), CategoryRequest{
		ID:          "test-id",
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error starting transaction")
	}
	mockDBConn.AssertExpectations(t)
}

// TestCategoryServiceImpl_UpdateCategory_CommitError tests the UpdateCategory method when the database transaction
// commit operation fails. This test verifies that the service properly handles commit errors and ensures
// that rollback is called when commit fails. It also checks that the appropriate error is returned.
func TestCategoryServiceImpl_UpdateCategory_CommitError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}
	mockTx := &MockCategoryDBTx{}

	// Mock successful BeginTx
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock successful UpdateCategories
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdateCategories", mock.Anything, mock.Anything).Return(nil)
	// Mock Commit to return an error
	mockTx.On("Commit").Return(fmt.Errorf("commit failed"))
	mockTx.On("Rollback").Return(nil)

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	err := service.UpdateCategory(context.Background(), CategoryRequest{
		ID:          "test-id",
		Name:        "Test Category",
		Description: "Test Description",
	})

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "commit_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error committing transaction")
	}
	mockDBConn.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// TestCategoryServiceImpl_DeleteCategory_TransactionError tests the DeleteCategory method when the database transaction
// fails to start. This test verifies that the service properly handles transaction initialization errors
// and returns the appropriate error with the correct error code and message.
func TestCategoryServiceImpl_DeleteCategory_TransactionError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}

	// Mock BeginTx to return an error
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return((*MockCategoryDBTx)(nil), fmt.Errorf("transaction failed"))

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	err := service.DeleteCategory(context.Background(), "test-id")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error starting transaction")
	}
	mockDBConn.AssertExpectations(t)
}

// TestCategoryServiceImpl_DeleteCategory_CommitError tests the DeleteCategory method when the database transaction
// commit operation fails. This test verifies that the service properly handles commit errors and ensures
// that rollback is called when commit fails. It also checks that the appropriate error is returned.
func TestCategoryServiceImpl_DeleteCategory_CommitError(t *testing.T) {
	mockDB := &MockCategoryDBQueries{}
	mockDBConn := &MockCategoryDBConn{}
	mockTx := &MockCategoryDBTx{}

	// Mock successful BeginTx
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock successful DeleteCategory
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("DeleteCategory", mock.Anything, "test-id").Return(nil)
	// Mock Commit to return an error
	mockTx.On("Commit").Return(fmt.Errorf("commit failed"))
	mockTx.On("Rollback").Return(nil)

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockDBConn,
	}

	err := service.DeleteCategory(context.Background(), "test-id")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		assert.Equal(t, "commit_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error committing transaction")
	}
	mockDBConn.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// TestCategoryService_DatabaseErrorScenarios tests the CreateCategory and UpdateCategory methods for database errors.
func TestCategoryService_DatabaseErrorScenarios(t *testing.T) {
	tests := []struct {
		name       string
		operation  string // "create" or "update"
		params     any
		setupMocks func(mockService *MockCategoryService)
		errorCode  string
	}{
		{
			name:      "CreateCategory_DatabaseError",
			operation: "create",
			params: CategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
			},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("", &handlers.AppError{Code: "create_category_error", Message: "database error"})
			},
			errorCode: "create_category_error",
		},
		{
			name:      "UpdateCategory_DatabaseError",
			operation: "update",
			params: CategoryRequest{
				ID:   "test-id",
				Name: "Test Category",
			},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(&handlers.AppError{Code: "update_category_error", Message: "database error"})
			},
			errorCode: "update_category_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCategoryService)
			tt.setupMocks(mockService)

			var err error
			switch tt.operation {
			case "create":
				_, err = mockService.CreateCategory(context.Background(), tt.params.(CategoryRequest))
			case "update":
				err = mockService.UpdateCategory(context.Background(), tt.params.(CategoryRequest))
			}
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tt.errorCode, appErr.Code)
		})
	}
}

// runCategoryServiceTest is a shared helper for testing category service create/update methods.
func runCategoryServiceTest(
	t *testing.T,
	serviceMethod func(*categoryServiceImpl, context.Context, CategoryRequest) (string, error),
	params CategoryRequest,
	setupMocks func(*MockCategoryDBQueries, *MockCategoryDBConn, *MockCategoryDBTx),
	expectedError bool,
	errorCode string,
) {
	mockDB := &MockCategoryDBQueries{}
	mockConn := &MockCategoryDBConn{}
	mockTx := &MockCategoryDBTx{}

	if setupMocks != nil {
		setupMocks(mockDB, mockConn, mockTx)
	}

	service := &categoryServiceImpl{
		db:     mockDB,
		dbConn: mockConn,
	}

	id, err := serviceMethod(service, context.Background(), params)

	if expectedError {
		require.Error(t, err)
		appErr := &handlers.AppError{}
		if errors.As(err, &appErr) {
			assert.Equal(t, errorCode, appErr.Code)
		}
		assert.Empty(t, id)
	} else {
		require.NoError(t, err)
		// Do not assert NotEmpty on id for UpdateCategory, since it always returns ""
	}

	mockDB.AssertExpectations(t)
	mockConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// Shared helper for database error setupMocks
func setupDatabaseErrorForCreate(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
	mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
	mockTx.On("Rollback").Return(nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("CreateCategory", mock.Anything, mock.Anything).Return(errors.New("database error"))
}

func setupDatabaseErrorForUpdate(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
	mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
	mockTx.On("Rollback").Return(nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdateCategories", mock.Anything, mock.Anything).Return(errors.New("database error"))
}
