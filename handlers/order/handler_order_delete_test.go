// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"context"
	"encoding/json"
	"errors"
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

// handler_order_delete_test.go: Tests for HandlerDeleteOrder covering all typical and edge cases.

const (
	testOrderID     = "order123"
	testNonexistent = "nonexistent"
)

// Helper to set chi URL param in request context
func setChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}

// TestHandlerDeleteOrder_Success verifies successful order deletion.
func TestHandlerDeleteOrder_Success(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID

	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "delete_order", "Deleted order successful", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order deleted successfully", response.Message)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteOrder_MissingOrderID checks that missing order ID returns an error.
func TestHandlerDeleteOrder_MissingOrderID(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	mockLogger.On("LogHandlerError", mock.Anything, "delete_order", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("DELETE", "/orders/", nil)
	// Do NOT set order_id param here
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Missing order_id", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "DeleteOrder")
}

// TestHandlerDeleteOrder_OrderNotFound checks that order not found returns the correct error.
func TestHandlerDeleteOrder_OrderNotFound(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testNonexistent

	appError := &handlers.AppError{Code: "order_not_found", Message: "Order not found"}
	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_order", "order_not_found", "Order not found", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order not found", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteOrder_DeleteFailed checks that delete failure returns the correct error.
func TestHandlerDeleteOrder_DeleteFailed(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID

	appError := &handlers.AppError{Code: "delete_order_error", Message: "Failed to delete order"}
	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_order", "delete_order_error", "Failed to delete order", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteOrder_Unauthorized checks that unauthorized access returns the correct error.
func TestHandlerDeleteOrder_Unauthorized(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID

	appError := &handlers.AppError{Code: "unauthorized", Message: "User is not authorized to delete this order"}
	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_order", "unauthorized", "User is not authorized to delete this order", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User is not authorized to delete this order", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteOrder_UnknownError checks that unknown errors are handled properly.
func TestHandlerDeleteOrder_UnknownError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID

	unknownError := errors.New("unknown database error")
	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_order", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteOrder_CompleteRequest verifies a complete delete request with all fields.
func TestHandlerDeleteOrder_CompleteRequest(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID

	mockOrderService.On("DeleteOrder", mock.Anything, orderID).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "delete_order", "Deleted order successful", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("DELETE", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "order_id", orderID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteOrder(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order deleted successfully", response.Message)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
