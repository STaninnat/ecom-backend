// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_category_update_integration_test.go: Integration tests for UpdateCategory HTTP handler with real logger and mock service.

// TestIntegration_HandlerUpdateCategory tests the update category handler with real logger and mock service.
// Covers successful updates, validation errors, service errors, and various edge cases.
func TestIntegration_HandlerUpdateCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		requestBody    any
		contentType    string
		user           database.User
		mockSetup      func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful update",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:          "test-category-id",
				Name:        "Updated Electronics",
				Description: "Updated description for electronics",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:          "test-category-id",
					Name:        "Updated Electronics",
					Description: "Updated description for electronics",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category updated successfully"}`,
		},
		{
			name:       "service error",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:          "test-category-id",
				Name:        "Updated Electronics",
				Description: "Updated description for electronics",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:          "test-category-id",
					Name:        "Updated Electronics",
					Description: "Updated description for electronics",
				}).Return(&handlers.AppError{Code: "update_category_error", Message: "Category not found"})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
		},
		{
			name:           "invalid JSON",
			categoryID:     "test-category-id",
			requestBody:    `{"invalid": json}`,
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(_ *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:           "missing content type",
			categoryID:     "test-category-id",
			requestBody:    `{"invalid": json}`,
			contentType:    "",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(_ *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:       "empty name",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:   "test-category-id",
				Name: "",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:   "test-category-id",
					Name: "",
				}).Return(&handlers.AppError{Code: "invalid_request", Message: "Category name is required"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name is required"}`,
		},
		{
			name:       "empty ID",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:   "",
				Name: "Updated Electronics",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:   "",
					Name: "Updated Electronics",
				}).Return(&handlers.AppError{Code: "invalid_request", Message: "Category ID is required"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name:       "name too long",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:   "test-category-id",
				Name: strings.Repeat("a", 101),
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:   "test-category-id",
					Name: strings.Repeat("a", 101),
				}).Return(&handlers.AppError{Code: "invalid_request", Message: "Category name too long (max 100 characters)"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name too long (max 100 characters)"}`,
		},
		{
			name:       "description too long",
			categoryID: "test-category-id",
			requestBody: CategoryRequest{
				ID:          "test-category-id",
				Name:        "Updated Electronics",
				Description: strings.Repeat("a", 501),
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:          "test-category-id",
					Name:        "Updated Electronics",
					Description: strings.Repeat("a", 501),
				}).Return(&handlers.AppError{Code: "invalid_request", Message: "Category description too long (max 500 characters)"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category description too long (max 500 characters)"}`,
		},
		{
			name:           "empty request body",
			categoryID:     "test-category-id",
			requestBody:    "",
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(_ *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:           "malformed json",
			categoryID:     "test-category-id",
			requestBody:    `{"id":"test-category-id","name":}`,
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(_ *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:        "service returns non-AppError",
			categoryID:  "test-category-id",
			requestBody: `{"id":"test-category-id","name":"Updated Category"}`,
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("UpdateCategory", mock.Anything, CategoryRequest{
					ID:   "test-category-id",
					Name: "Updated Category",
				}).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &MockCategoryService{}
			tt.mockSetup(mockService)

			// Create a real logger for integration testing
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel) // Only log errors to reduce noise

			// Create the config with proper logger setup
			cfg := &HandlersCategoryConfig{
				Config: &handlers.Config{
					Logger: logger,
				},
			}
			// Set the Logger field to the embedded config which implements HandlerLogger
			cfg.Logger = cfg.Config

			// Set the mock service
			cfg.categoryService = mockService

			// Create request body
			var body []byte
			var err error
			if strBody, ok := tt.requestBody.(string); ok {
				body = []byte(strBody)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/categories/"+tt.categoryID, bytes.NewBuffer(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			req.Header.Set("User-Agent", "test-agent")

			// Set the URL parameter for chi router
			muxSetURLParam(req, "id", tt.categoryID)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			cfg.HandlerUpdateCategory(w, req, tt.user)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body
			responseBody := strings.TrimSpace(w.Body.String())
			assert.Equal(t, tt.expectedBody, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
