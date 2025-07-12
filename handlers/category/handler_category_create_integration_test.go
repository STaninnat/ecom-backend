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

func TestIntegration_HandlerCreateCategory(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		contentType    string
		user           database.User
		mockSetup      func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful_creation",
			requestBody: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices and gadgets",
			},
			contentType: "application/json",
			user: database.User{
				ID: "test-user-id",
			},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "Electronics",
					Description: "Electronic devices and gadgets",
				}).Return("new-category-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name: "service_error",
			requestBody: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices and gadgets",
			},
			contentType: "application/json",
			user: database.User{
				ID: "test-user-id",
			},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "Electronics",
					Description: "Electronic devices and gadgets",
				}).Return("", &handlers.AppError{
					Code:    "database_error",
					Message: "Database connection failed",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
		},
		{
			name:           "invalid_json",
			requestBody:    `{"name": "Electronics", "description": "Invalid JSON`,
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(mockService *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name: "missing_content_type",
			requestBody: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices and gadgets",
			},
			contentType: "",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "Electronics",
					Description: "Electronic devices and gadgets",
				}).Return("new-category-id", nil)
			},
			expectedStatus: http.StatusCreated, // Handler doesn't check content-type
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name: "empty_user",
			requestBody: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices and gadgets",
			},
			contentType: "application/json",
			user:        database.User{},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "Electronics",
					Description: "Electronic devices and gadgets",
				}).Return("new-category-id", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Category created successfully"}`,
		},
		{
			name:        "service_returns_non_AppError",
			requestBody: `{"name":"Test Category"}`,
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name: "Test Category",
				}).Return("", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name:           "empty_request_body",
			requestBody:    "",
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(mockService *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:           "malformed_json",
			requestBody:    `{"name":}`,
			contentType:    "application/json",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(mockService *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request payload"}`,
		},
		{
			name:        "category_already_exists_error",
			requestBody: `{"name":"Test Category","description":"Test Description"}`,
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "Test Category",
					Description: "Test Description",
				}).Return("", &handlers.AppError{
					Code:    "create_category_error",
					Message: "Category already exists",
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
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
				HandlersConfig: &handlers.HandlersConfig{
					Logger: logger,
				},
			}
			// Set the Logger field to the embedded config which implements HandlerLogger
			cfg.Logger = cfg.HandlersConfig

			// Set the mock service
			cfg.categoryService = mockService

			// Create request body
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("X-Forwarded-For", "192.0.2.1")
			req.Header.Set("User-Agent", "test-agent")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			cfg.HandlerCreateCategory(w, req, tt.user)

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

func TestIntegration_HandlerCreateCategory_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		contentType    string
		user           database.User
		mockSetup      func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "name_too_long",
			requestBody: CategoryRequest{
				Name:        strings.Repeat("a", 101),
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        strings.Repeat("a", 101),
					Description: "desc",
				}).Return("", &handlers.AppError{
					Code:    "invalid_request",
					Message: "Category name too long",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name too long"}`,
		},
		{
			name: "empty_name",
			requestBody: CategoryRequest{
				Name:        "",
				Description: "desc",
			},
			contentType: "application/json",
			user:        database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("CreateCategory", mock.Anything, CategoryRequest{
					Name:        "",
					Description: "desc",
				}).Return("", &handlers.AppError{
					Code:    "invalid_request",
					Message: "Category name is required",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category name is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &MockCategoryService{}
			tt.mockSetup(mockService)

			// Create a real logger for integration testing
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			// Create the config with proper logger setup
			cfg := &HandlersCategoryConfig{
				HandlersConfig: &handlers.HandlersConfig{
					Logger: logger,
				},
			}
			cfg.Logger = cfg.HandlersConfig
			cfg.categoryService = mockService

			// Create request body
			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			cfg.HandlerCreateCategory(w, req, tt.user)

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
