// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_product_delete_test.go: Tests the delete product handler for success, missing ID, and service error with expected responses and logging.

// TestHandlerDeleteProduct_Success tests the successful deletion of a product via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service deletes the product without error.
func TestHandlerDeleteProduct_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	mockService.On("DeleteProduct", mock.Anything, "pid1").Return(nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "delete_product", "Delete success", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("DELETE", "/products/pid1", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "pid1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	w := httptest.NewRecorder()

	cfg.HandlerDeleteProduct(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.HandlerResponse
	_ = resp
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerDeleteProduct_MissingID tests the handler's response when the product ID is missing from the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerDeleteProduct_MissingID(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	mockLog.On("LogHandlerError", mock.Anything, "delete_product", "invalid_request", "Product ID is required", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("DELETE", "/products/", nil)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteProduct(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerDeleteProduct_ServiceError tests the handler's behavior when the product service returns an error during deletion.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerDeleteProduct_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	err := &handlers.AppError{Code: "delete_product_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("DeleteProduct", mock.Anything, "pid1").Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "delete_product", "delete_product_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("DELETE", "/products/pid1", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "pid1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	w := httptest.NewRecorder()

	cfg.HandlerDeleteProduct(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
