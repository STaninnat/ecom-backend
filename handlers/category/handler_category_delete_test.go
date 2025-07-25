// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_delete_test.go: Tests for HandlerDeleteCategory with various scenarios and input validations

// TestHandlerDeleteCategory tests the delete category handler with mock service and logger.
// Covers successful deletion, missing parameters, and service errors.
func TestHandlerDeleteCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		setupMocks     func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful deletion",
			categoryID: "test-id",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category deleted successfully"}`,
		},
		{
			name:           "missing category ID",
			categoryID:     "",
			setupMocks:     func(_ *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name:       "service error",
			categoryID: "test-id",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-id").Return(&handlers.AppError{
					Code:    "delete_category_error",
					Message: "Category not found",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCategoryService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			testConfig := &TestHandlersCategoryConfig{
				MockHandlersConfig: &MockHandlersConfig{},
				Logger:             nil, // will set below
				categoryService:    mockService,
			}
			testConfig.Logger = testConfig.MockHandlersConfig
			testConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			testConfig.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryID, nil)
			if tt.categoryID != "" {
				req = muxSetURLParam(req, "id", tt.categoryID)
			}

			w := httptest.NewRecorder()

			user := database.User{
				ID:   "test-user-id",
				Name: "Test User",
			}

			testConfig.HandlerDeleteCategory(w, req, user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

// TestHandlerDeleteCategory_EdgeCases tests edge cases for the delete category handler.
// Covers invalid formats, empty users, missing parameters, and special characters.
func TestHandlerDeleteCategory_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		setupMocks     func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
		user           database.User
		setupRequest   func(*http.Request)
	}{
		{
			name:       "service returns non-AppError",
			categoryID: "test-id",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-id").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
			user:           database.User{ID: "test-user-id"},
		},
		{
			name:       "invalid category ID format",
			categoryID: "invalid-format-123",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "invalid-format-123").Return(&handlers.AppError{
					Code:    "delete_category_error",
					Message: "Invalid category ID format",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
			user:           database.User{ID: "test-user-id"},
		},
		{
			name:       "user is empty struct",
			categoryID: "test-id",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category deleted successfully"}`,
			user:           database.User{},
		},
		{
			name:       "URL parameter not set",
			categoryID: "",
			setupMocks: func(_ *MockCategoryService) {},
			setupRequest: func(_ *http.Request) {
				// Don't set the path value, simulating missing URL parameter
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
			user:           database.User{ID: "test-user-id"},
		},
		{
			name:       "very long category ID",
			categoryID: strings.Repeat("a", 1000),
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, strings.Repeat("a", 1000)).Return(&handlers.AppError{
					Code:    "delete_category_error",
					Message: "Category ID too long",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
			user:           database.User{ID: "test-user-id"},
		},
		{
			name:       "special characters in category ID",
			categoryID: "test-id-with-special-chars",
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-id-with-special-chars").Return(&handlers.AppError{
					Code:    "delete_category_error",
					Message: "Category not found",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
			user:           database.User{ID: "test-user-id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCategoryService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			testConfig := &TestHandlersCategoryConfig{
				MockHandlersConfig: &MockHandlersConfig{},
				Logger:             nil, // will set below
				categoryService:    mockService,
			}
			testConfig.Logger = testConfig.MockHandlersConfig
			testConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			testConfig.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryID, nil)
			if tt.categoryID != "" && tt.setupRequest == nil {
				req = muxSetURLParam(req, "id", tt.categoryID)
			}
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			w := httptest.NewRecorder()

			testConfig.HandlerDeleteCategory(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
