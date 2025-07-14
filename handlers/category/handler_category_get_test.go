package categoryhandlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerGetAllCategories tests the get all categories handler with mock service and logger.
// Covers successful retrieval with and without user, and service errors.
func TestHandlerGetAllCategories(t *testing.T) {
	tests := []struct {
		name           string
		user           *database.User
		setupMocks     func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful retrieval with user",
			user: &database.User{
				ID:   "test-user-id",
				Name: "Test User",
			},
			setupMocks: func(mockService *MockCategoryService) {
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
				mockService.On("GetAllCategories", mock.Anything).Return(expectedCategories, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"cat1","name":"Category 1","description":"Description 1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"},{"id":"cat2","name":"Category 2","description":"Description 2","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]`,
		},
		{
			name: "successful retrieval without user",
			user: nil,
			setupMocks: func(mockService *MockCategoryService) {
				expectedCategories := []database.Category{
					{
						ID:          "cat1",
						Name:        "Category 1",
						Description: utils.ToNullString("Description 1"),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				mockService.On("GetAllCategories", mock.Anything).Return(expectedCategories, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"cat1","name":"Category 1","description":"Description 1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]`,
		},
		{
			name: "service error",
			user: &database.User{
				ID:   "test-user-id",
				Name: "Test User",
			},
			setupMocks: func(mockService *MockCategoryService) {
				mockService.On("GetAllCategories", mock.Anything).Return([]database.Category{}, &handlers.AppError{
					Code:    "database_error",
					Message: "Database connection failed",
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

			req := httptest.NewRequest("GET", "/categories", nil)
			w := httptest.NewRecorder()

			testConfig.HandlerGetAllCategories(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), `"id":"cat1"`)
				assert.Contains(t, w.Body.String(), `"name":"Category 1"`)
			} else {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
