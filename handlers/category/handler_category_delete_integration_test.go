package categoryhandlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestIntegration_HandlerDeleteCategory tests the delete category handler with real logger and mock service.
// Covers successful deletion, service errors, missing parameters, and edge cases.
func TestIntegration_HandlerDeleteCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		user           database.User
		mockSetup      func(*MockCategoryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful deletion",
			categoryID: "test-category-id",
			user:       database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-category-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Category deleted successfully"}`,
		},
		{
			name:       "service error",
			categoryID: "test-category-id",
			user:       database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-category-id").Return(&handlers.AppError{Code: "delete_category_error", Message: "Category not found"})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong, please try again later"}`,
		},
		{
			name:           "missing category ID",
			categoryID:     "",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(mockService *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name:           "empty category ID",
			categoryID:     "",
			user:           database.User{ID: "test-user-id"},
			mockSetup:      func(mockService *MockCategoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Category ID is required"}`,
		},
		{
			name:       "service returns non-AppError",
			categoryID: "test-category-id",
			user:       database.User{ID: "test-user-id"},
			mockSetup: func(mockService *MockCategoryService) {
				mockService.On("DeleteCategory", mock.Anything, "test-category-id").Return(assert.AnError)
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
				HandlersConfig: &handlers.HandlersConfig{
					Logger: logger,
				},
			}
			// Set the Logger field to the embedded config which implements HandlerLogger
			cfg.Logger = cfg.HandlersConfig

			// Set the mock service
			cfg.categoryService = mockService

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/categories/"+tt.categoryID, nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			req.Header.Set("User-Agent", "test-agent")

			// Set up chi router context with URL parameter if categoryID is provided
			if tt.categoryID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.categoryID)
				ctx := req.Context()
				ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			cfg.HandlerDeleteCategory(w, req, tt.user)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// muxSetURLParam sets a chi URL param for the request (for testing)
func muxSetURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := r.Context()
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}
