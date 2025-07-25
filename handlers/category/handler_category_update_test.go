// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_update_test.go: Tests for UpdateCategory HTTP handler with various input and error scenarios.

// TestHandlerUpdateCategory tests the update category handler with mock service and logger.
// Covers successful updates, invalid requests, and service errors.
func TestHandlerUpdateCategory(t *testing.T) {

	tests := []struct {
		name           string
		requestBody    any
		contentType    string
		user           database.User
		setupMocks     func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful update",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Updated Category",
				Description: "Updated Description",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:          "test-id",
					Name:        "Updated Category",
					Description: "Updated Description",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name: "invalid request body",
			requestBody: CategoryRequest{
				Name: "Test Category",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					Name: "Test Category",
				}).Return(&handlers.AppError{
					Code:    "invalid_request",
					Message: "Category ID is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name: "service error",
			requestBody: CategoryRequest{
				ID:   "test-id",
				Name: "Test Category",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(&handlers.AppError{
					Code:    "invalid_request",
					Message: "Category ID is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
	}

	runCategoryHandlerTestTable(t, tests, "PUT", func(cfg *TestHandlersCategoryConfig, w http.ResponseWriter, req *http.Request, user database.User) {
		cfg.HandlerUpdateCategory(w, req, user)
	})
}

// TestHandlerUpdateCategory_InvalidJSON tests the update category handler with malformed JSON.
// Verifies proper error handling for invalid request payloads.
func TestHandlerUpdateCategory_InvalidJSON(t *testing.T) {
	mockService := &MockCategoryService{}
	testConfig := &TestHandlersCategoryConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Logger:             nil, // will set below
		categoryService:    mockService,
	}
	testConfig.Logger = testConfig.MockHandlersConfig
	testConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("PUT", "/categories", bytes.NewBufferString(`{"id": "test"`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	user := database.User{
		ID:   "test-user-id",
		Name: "Test User",
	}
	testConfig.HandlerUpdateCategory(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid request payload"}`, w.Body.String())
}

// TestHandlerUpdateCategory_EdgeCases tests edge cases for the update category handler.
// Covers validation errors, content types, empty users, and various input scenarios.
func TestHandlerUpdateCategory_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		contentType    string
		user           database.User
		setupMocks     func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid update",
			requestBody: CategoryRequest{ID: "cat-1", Name: "Updated Category"},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name: "name too long",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        strings.Repeat("a", 101),
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(&handlers.AppError{
					Code:    "invalid_request",
					Message: "Category name too long (max 100 characters)",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name too long (max 100 characters)"}`,
		},
		{
			name: "description too long",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Valid Name",
				Description: strings.Repeat("a", 501),
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(&handlers.AppError{
					Code:    "invalid_request",
					Message: "Category description too long (max 500 characters)",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category description too long (max 500 characters)"}`,
		},
		{
			name: "service returns non-AppError",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name: "missing content-type",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name: "incorrect content-type",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "text/plain",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name: "user is empty struct",
			requestBody: CategoryRequest{
				ID:          "test-id",
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name: "ID is empty",
			requestBody: CategoryRequest{
				ID:          "",
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(&handlers.AppError{
					Code:    "invalid_request",
					Message: "Category ID is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name: "ID is very long",
			requestBody: CategoryRequest{
				ID:          strings.Repeat("a", 1000),
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
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

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/categories", bytes.NewBuffer(requestBodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			testConfig.HandlerUpdateCategory(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
