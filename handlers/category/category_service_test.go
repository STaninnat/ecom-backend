package categoryhandlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("CreateCategory", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: true,
			errorCode:     "create_category_error",
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

			id, err := service.CreateCategory(context.Background(), tt.params)

			if tt.expectedError {
				assert.Error(t, err)
				if appErr, ok := err.(*handlers.AppError); ok {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, id)
			}

			mockDB.AssertExpectations(t)
			mockConn.AssertExpectations(t)
			mockTx.AssertExpectations(t)
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
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
			setupMocks: func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {
				mockConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockDB.On("WithTx", mockTx).Return(mockDB)
				mockDB.On("UpdateCategories", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: true,
			errorCode:     "update_category_error",
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

			err := service.UpdateCategory(context.Background(), tt.params)

			if tt.expectedError {
				assert.Error(t, err)
				if appErr, ok := err.(*handlers.AppError); ok {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockConn.AssertExpectations(t)
			mockTx.AssertExpectations(t)
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
			setupMocks:    func(mockDB *MockCategoryDBQueries, mockConn *MockCategoryDBConn, mockTx *MockCategoryDBTx) {},
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
				assert.Error(t, err)
				if appErr, ok := err.(*handlers.AppError); ok {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
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
				assert.Error(t, err)
				if appErr, ok := err.(*handlers.AppError); ok {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
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

	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
		assert.Equal(t, "commit_error", appErr.Code)
		assert.Contains(t, appErr.Message, "Error committing transaction")
	}
	mockDBConn.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
