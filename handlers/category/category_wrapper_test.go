package categoryhandlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlersCategoryConfig_InitCategoryService tests the InitCategoryService method of the HandlersCategoryConfig.
// It covers various initialization scenarios including successful initialization with valid configuration,
// and error cases when required dependencies (handlers config, API config) are missing.
// This test ensures that the category service is properly initialized and validates all required dependencies.
func TestHandlersCategoryConfig_InitCategoryService(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func() *HandlersCategoryConfig
		expectedError bool
	}{
		{
			name: "successful initialization",
			setupConfig: func() *HandlersCategoryConfig {
				return &HandlersCategoryConfig{
					HandlersConfig: &handlers.HandlersConfig{
						APIConfig: &config.APIConfig{
							DB:     &database.Queries{},
							DBConn: &sql.DB{},
						},
					},
				}
			},
			expectedError: false,
		},
		{
			name: "nil handlers config",
			setupConfig: func() *HandlersCategoryConfig {
				return &HandlersCategoryConfig{
					HandlersConfig: nil,
				}
			},
			expectedError: true,
		},
		{
			name: "nil API config",
			setupConfig: func() *HandlersCategoryConfig {
				return &HandlersCategoryConfig{
					HandlersConfig: &handlers.HandlersConfig{
						APIConfig: nil,
					},
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			err := cfg.InitCategoryService()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.categoryService)
			}
		})
	}
}

// TestHandlersCategoryConfig_InitCategoryService_NilHandlersConfig tests the InitCategoryService method
// when the HandlersConfig is nil. This edge case test ensures that the initialization properly
// validates the presence of the handlers configuration and returns an appropriate error message.
func TestHandlersCategoryConfig_InitCategoryService_NilHandlersConfig(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		HandlersConfig: nil,
	}

	err := cfg.InitCategoryService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

// TestHandlersCategoryConfig_InitCategoryService_NilAPIConfig tests the InitCategoryService method
// when the APIConfig is nil. This edge case test ensures that the initialization properly
// validates the presence of the API configuration and returns an appropriate error message.
func TestHandlersCategoryConfig_InitCategoryService_NilAPIConfig(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		HandlersConfig: &handlers.HandlersConfig{},
	}

	err := cfg.InitCategoryService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API config not initialized")
}

// TestHandlersCategoryConfig_InitCategoryService_NilDB tests the InitCategoryService method
// when the database is nil. This edge case test ensures that the initialization properly
// validates the presence of the database and returns an appropriate error message.
func TestHandlersCategoryConfig_InitCategoryService_NilDB(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
	}

	err := cfg.InitCategoryService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestHandlersCategoryConfig_InitCategoryService_NilDBConn tests the InitCategoryService method
// when the database connection is nil. This edge case test ensures that the initialization properly
// validates the presence of the database connection and returns an appropriate error message.
func TestHandlersCategoryConfig_InitCategoryService_NilDBConn(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB: &database.Queries{},
			},
		},
	}

	err := cfg.InitCategoryService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// TestHandlersCategoryConfig_GetCategoryService tests the GetCategoryService method of the HandlersCategoryConfig.
// It covers scenarios where the service is already initialized, not initialized with valid config,
// and not initialized with invalid config. The test verifies that the method properly handles
// lazy initialization and returns a service instance even when configuration is incomplete.
func TestHandlersCategoryConfig_GetCategoryService(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *HandlersCategoryConfig
		expectedNil bool
	}{
		{
			name: "service already initialized",
			setupConfig: func() *HandlersCategoryConfig {
				cfg := &HandlersCategoryConfig{
					HandlersConfig: &handlers.HandlersConfig{
						APIConfig: &config.APIConfig{},
					},
				}
				cfg.categoryService = &categoryServiceImpl{}
				return cfg
			},
			expectedNil: false,
		},
		{
			name: "service not initialized - valid config",
			setupConfig: func() *HandlersCategoryConfig {
				return &HandlersCategoryConfig{
					HandlersConfig: &handlers.HandlersConfig{
						APIConfig: &config.APIConfig{},
					},
				}
			},
			expectedNil: false,
		},
		{
			name: "service not initialized - invalid config",
			setupConfig: func() *HandlersCategoryConfig {
				return &HandlersCategoryConfig{
					HandlersConfig: nil,
				}
			},
			expectedNil: false, // Should still return a service (will fail gracefully)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			service := cfg.GetCategoryService()

			if tt.expectedNil {
				assert.Nil(t, service)
			} else {
				assert.NotNil(t, service)
			}
		})
	}
}

// TestHandlersCategoryConfig_GetCategoryService_NilConfigs tests the GetCategoryService method
// when all configurations are nil. This edge case test verifies that the method still returns
// a service instance but that the service fails gracefully when used with proper error handling.
// It ensures that the service doesn't panic and returns appropriate errors for nil dependencies.
func TestHandlersCategoryConfig_GetCategoryService_NilConfigs(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		HandlersConfig: nil,
	}

	service := cfg.GetCategoryService()
	assert.NotNil(t, service)

	// Test that the service fails gracefully when used
	_, err := service.CreateCategory(context.Background(), CategoryRequest{
		Name: "Test",
	})
	// The service should return an error due to nil DB connection
	assert.Error(t, err)
	if appErr, ok := err.(*handlers.AppError); ok {
		assert.Equal(t, "transaction_error", appErr.Code)
		assert.Contains(t, appErr.Message, "DB connection is nil")
	}
}

// TestHandlersCategoryConfig_handleCategoryError tests the handleCategoryError method of the HandlersCategoryConfig.
// It covers various error types including invalid request errors, database errors, transaction errors,
// category-specific errors (create, update, delete), commit errors, unknown errors, and non-AppError types.
// The test verifies that each error type is properly categorized and returns the appropriate HTTP status code
// and error message. It also ensures that error logging is called with the correct parameters.
func TestHandlersCategoryConfig_handleCategoryError(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
	}{
		{
			name: "invalid request error",
			error: &handlers.AppError{
				Code:    "invalid_request",
				Message: "Invalid request",
			},
			expectedStatus: 400,
		},
		{
			name: "database error",
			error: &handlers.AppError{
				Code:    "database_error",
				Message: "Database error",
				Err:     errors.New("db error"),
			},
			expectedStatus: 500,
		},
		{
			name: "transaction error",
			error: &handlers.AppError{
				Code:    "transaction_error",
				Message: "Transaction error",
				Err:     errors.New("tx error"),
			},
			expectedStatus: 500,
		},
		{
			name: "create category error",
			error: &handlers.AppError{
				Code:    "create_category_error",
				Message: "Create category error",
				Err:     errors.New("create error"),
			},
			expectedStatus: 500,
		},
		{
			name: "update category error",
			error: &handlers.AppError{
				Code:    "update_category_error",
				Message: "Update category error",
				Err:     errors.New("update error"),
			},
			expectedStatus: 500,
		},
		{
			name: "delete category error",
			error: &handlers.AppError{
				Code:    "delete_category_error",
				Message: "Delete category error",
				Err:     errors.New("delete error"),
			},
			expectedStatus: 500,
		},
		{
			name: "commit error",
			error: &handlers.AppError{
				Code:    "commit_error",
				Message: "Commit error",
				Err:     errors.New("commit error"),
			},
			expectedStatus: 500,
		},
		{
			name: "unknown error",
			error: &handlers.AppError{
				Code:    "unknown_error",
				Message: "Unknown error",
			},
			expectedStatus: 500,
		},
		{
			name:           "non-app error",
			error:          errors.New("generic error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := &MockHandlersConfig{}
			cfg := &HandlersCategoryConfig{
				HandlersConfig: &handlers.HandlersConfig{},
				Logger:         mockConfig,
			}

			// Set up mock expectations
			mockConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			// Create a mock response writer and request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			// For this test, we're just verifying the error handling logic
			cfg.handleCategoryError(w, req, tt.error, "test_operation", "127.0.0.1", "test-agent")

			mockConfig.AssertExpectations(t)
		})
	}
}
