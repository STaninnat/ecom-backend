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

// handler_category_create_test.go: Tests for the category creation HTTP handler, covering success and error cases.

// TestHandlerCreateCategory tests the category creation handler with basic scenarios including
// successful creation, invalid request body, and service errors using mocked dependencies.
func TestHandlerCreateCategory(t *testing.T) {
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
			name:        "Valid request",
			requestBody: CategoryRequest{Name: "Test Category", Description: "Test Description"},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("test-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name: "invalid request body",
			requestBody: CategoryRequest{
				Name: "", // Empty name should trigger validation error
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name: "",
				}).Return("", &handlers.AppError{
					Code:    "invalid_request",
					Message: "Category name is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name is required"}`,
		},
		{
			name: "service error",
			requestBody: CategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("", &handlers.AppError{
					Code:    "invalid_request",
					Message: "Category name is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name is required"}`,
		},
	}

	runCategoryHandlerTestTable(t, tests, "POST", func(cfg *TestHandlersCategoryConfig, w http.ResponseWriter, req *http.Request, user database.User) {
		cfg.HandlerCreateCategory(w, req, user)
	})
}

// TestHandlerCreateCategory_InvalidJSON tests the handler's response when invalid JSON is provided
// in the request body, ensuring proper error handling and response formatting.
func TestHandlerCreateCategory_InvalidJSON(t *testing.T) {
	// Create test config that embeds the mock
	testConfig := &TestHandlersCategoryConfig{
		MockHandlersConfig: &MockHandlersConfig{},
		Logger:             nil, // will set below
		categoryService:    &MockCategoryService{},
	}
	testConfig.Logger = testConfig.MockHandlersConfig
	testConfig.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(`{"name": "Test"`))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create test user
	user := database.User{
		ID:   "test-user-id",
		Name: "Test User",
	}

	// Call handler
	testConfig.HandlerCreateCategory(w, req, user)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid request payload"}`, w.Body.String())
}

// TestHandlerCreateCategory_EdgeCases tests various edge cases for category creation including
// validation errors, service errors, content-type handling, and different user states.
func TestHandlerCreateCategory_EdgeCases(t *testing.T) {
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
			name: "name too long",
			requestBody: CategoryRequest{
				Name:        strings.Repeat("a", 101),
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("", &handlers.AppError{
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
				Name:        "Valid Name",
				Description: strings.Repeat("a", 501),
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("", &handlers.AppError{
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
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name: "missing content-type",
			requestBody: CategoryRequest{
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("test-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name: "incorrect content-type",
			requestBody: CategoryRequest{
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "text/plain",
			user:        database.User{ID: "test-user-id"},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("test-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name: "user is empty struct",
			requestBody: CategoryRequest{
				Name:        "Valid Name",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, mock.Anything).Return("test-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
	}

	runCategoryHandlerTestTable(t, tests, "POST", func(cfg *TestHandlersCategoryConfig, w http.ResponseWriter, req *http.Request, user database.User) {
		cfg.HandlerCreateCategory(w, req, user)
	})
}

// Shared helper for category handler test table loops
func runCategoryHandlerTestTable(t *testing.T, tests []struct {
	name           string
	requestBody    any
	contentType    string
	user           database.User
	setupMocks     func(*MockCategoryService)
	expectedStatus int
	expectedBody   string
}, method string, handlerFunc func(cfg *TestHandlersCategoryConfig, w http.ResponseWriter, req *http.Request, user database.User)) {
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
			req := httptest.NewRequest(method, "/categories", bytes.NewBuffer(requestBodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			handlerFunc(testConfig, w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
