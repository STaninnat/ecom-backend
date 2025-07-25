// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_category_get_integration_test.go: Integration tests covering success and error scenarios of GetAllCategories handler.

// TestIntegration_HandlerGetAllCategories tests the get all categories handler with real logger and mock service.
// Covers successful retrieval and service errors with proper response validation.
func TestIntegration_HandlerGetAllCategories(t *testing.T) {
	cfg := &HandlersCategoryConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockService := &MockCategoryServiceForGetIntegration{}
		mockLogger := &MockLoggerForGetIntegration{}
		cfg.categoryService = mockService
		cfg.Logger = mockLogger
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedCategories := []database.Category{
			{
				ID:          "cat1",
				Name:        "Category 1",
				Description: utils.ToNullString("Description 1"),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}
		mockService.On("GetAllCategories", mock.Anything).Return(expectedCategories, nil)

		mockLogger.On(
			"LogHandlerSuccess",
			mock.Anything, // context
			"get_all_categories",
			"Categories fetched successfully",
			mock.Anything, // ip
			mock.Anything, // user agent
		).Return()

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()
		user := &database.User{ID: "test-user-id", Name: "Test User"}

		cfg.HandlerGetAllCategories(w, req, user)

		assert.Equal(t, http.StatusOK, w.Code)
		var got []map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		assert.Equal(t, "cat1", got[0]["ID"])
		assert.Equal(t, "Category 1", got[0]["Name"])
		description := got[0]["Description"].(map[string]any)
		assert.Equal(t, "Description 1", description["String"])
		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := &MockCategoryServiceForGetIntegration{}
		mockLogger := &MockLoggerForGetIntegration{}
		cfg.categoryService = mockService
		cfg.Logger = mockLogger
		mockService.On("GetAllCategories", mock.Anything).Return([]database.Category{}, &handlers.AppError{
			Code:    "database_error",
			Message: "Database connection failed",
		})
		mockLogger.On("LogHandlerError", mock.Anything, "get_all_categories", "database_error", "Database connection failed", mock.Anything, mock.Anything, mock.Anything).Return()

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()
		user := &database.User{ID: "test-user-id", Name: "Test User"}

		cfg.HandlerGetAllCategories(w, req, user)

		println("DEBUG: status=", w.Code, "body=", w.Body.String())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error":"Something went wrong, please try again later"}`, w.Body.String())
		mockService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
