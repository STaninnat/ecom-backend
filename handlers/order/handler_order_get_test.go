// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_order_get_test.go: Tests for order handlers, verifying behavior for admin and user endpoints including success, validation, and error handling.

// TestHandlerGetAllOrders_Success verifies successful retrieval of all orders.
func TestHandlerGetAllOrders_Success(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	expectedOrders := []database.Order{
		{ID: "order1", UserID: "user1", TotalAmount: "100.00", Status: "pending"},
		{ID: "order2", UserID: "user2", TotalAmount: "200.00", Status: "completed"},
	}

	mockOrderService.On("GetAllOrders", mock.Anything).Return(expectedOrders, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "list_all_orders", "Listed all orders", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllOrders(w, req, database.User{})

	assert.Equal(t, http.StatusOK, w.Code)
	var response []database.Order
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetAllOrders_ServiceError checks that service errors are handled properly.
func TestHandlerGetAllOrders_ServiceError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	appError := &handlers.AppError{Code: "database_error", Message: "Database connection failed"}
	mockOrderService.On("GetAllOrders", mock.Anything).Return(([]database.Order)(nil), appError)
	mockLogger.On("LogHandlerError", mock.Anything, "list_all_orders", "internal_error", "Database connection failed", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllOrders(w, req, database.User{})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUserOrders_Success verifies successful retrieval of user orders.
func TestHandlerGetUserOrders_Success(t *testing.T) {
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
	expectedOrders := []UserOrderResponse{
		{OrderID: "order1", TotalAmount: "100.00", Status: "pending"},
		{OrderID: "order2", TotalAmount: "200.00", Status: "completed"},
	}

	mockOrderService.On("GetUserOrders", mock.Anything, user).Return(expectedOrders, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_user_orders", "Retrieved user orders", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/user/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetUserOrders(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []UserOrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUserOrders_ServiceError checks that service errors are handled properly.
func TestHandlerGetUserOrders_ServiceError(t *testing.T) {
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
	appError := &handlers.AppError{Code: "database_error", Message: "Database connection failed"}
	mockOrderService.On("GetUserOrders", mock.Anything, user).Return(([]UserOrderResponse)(nil), appError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_user_orders", "internal_error", "Database connection failed", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/user/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetUserOrders(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderByID_Success verifies successful retrieval of a specific order.
func TestHandlerGetOrderByID_Success(t *testing.T) {
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
	expectedOrder := &OrderDetailResponse{
		Order: database.Order{ID: orderID, UserID: user.ID, TotalAmount: "100.00", Status: "pending"},
		Items: []database.OrderItem{},
	}

	req := httptest.NewRequest("GET", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "orderID", orderID)

	mockOrderService.On("GetOrderByID", mock.Anything, orderID, user).Return(expectedOrder, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_order_by_id", "Fetched order details", mock.Anything, mock.Anything).Return()

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderByID(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response OrderDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, orderID, response.Order.ID)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderByID_MissingOrderID checks that missing order ID returns an error.
func TestHandlerGetOrderByID_MissingOrderID(t *testing.T) {
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
	mockLogger.On("LogHandlerError", mock.Anything, "get_order_by_id", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/orders/", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetOrderByID(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Missing order ID", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "GetOrderByID")
}

// TestHandlerGetOrderByID_OrderNotFound checks that order not found returns the correct error.
func TestHandlerGetOrderByID_OrderNotFound(t *testing.T) {
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
	orderID := "nonexistent"
	appError := &handlers.AppError{Code: "order_not_found", Message: "Order not found"}
	mockOrderService.On("GetOrderByID", mock.Anything, orderID, user).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_order_by_id", "order_not_found", "Order not found", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "orderID", orderID)

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderByID(w, req, user)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order not found", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderItemsByOrderID_Success verifies successful retrieval of order items.
func TestHandlerGetOrderItemsByOrderID_Success(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	orderID := testOrderID
	expectedItems := []OrderItemResponse{
		{ID: "item1", ProductID: "prod1", Quantity: 2, Price: "10.50"},
		{ID: "item2", ProductID: "prod2", Quantity: 1, Price: "25.00"},
	}

	mockOrderService.On("GetOrderItemsByOrderID", mock.Anything, orderID).Return(expectedItems, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_order_items", "Retrieved order items", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/orders/"+orderID+"/items", nil)
	req = setChiURLParam(req, "order_id", orderID)

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderItemsByOrderID(w, req, database.User{})

	assert.Equal(t, http.StatusOK, w.Code)
	var response []OrderItemResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderItemsByOrderID_MissingOrderID checks that missing order ID returns an error.
func TestHandlerGetOrderItemsByOrderID_MissingOrderID(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	mockLogger.On("LogHandlerError", mock.Anything, "get_order_items", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/orders//items", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetOrderItemsByOrderID(w, req, database.User{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Missing order_id", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "GetOrderItemsByOrderID")
}

// TestHandlerGetOrderItemsByOrderID_ServiceError checks that service errors are handled properly.
func TestHandlerGetOrderItemsByOrderID_ServiceError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	orderID := testOrderID
	appError := &handlers.AppError{Code: "database_error", Message: "Database connection failed"}
	mockOrderService.On("GetOrderItemsByOrderID", mock.Anything, orderID).Return(([]OrderItemResponse)(nil), appError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_order_items", "internal_error", "Database connection failed", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("GET", "/orders/"+orderID+"/items", nil)
	req = setChiURLParam(req, "order_id", orderID)

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderItemsByOrderID(w, req, database.User{})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetAllOrders_UnknownError checks that unknown errors are handled properly.
func TestHandlerGetAllOrders_UnknownError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	unknownError := errors.New("unknown database error")
	mockOrderService.On("GetAllOrders", mock.Anything).Return(([]database.Order)(nil), unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "list_all_orders", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllOrders(w, req, database.User{})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUserOrders_UnknownError checks that unknown errors are handled properly.
func TestHandlerGetUserOrders_UnknownError(t *testing.T) {
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
	unknownError := errors.New("unknown database error")
	mockOrderService.On("GetUserOrders", mock.Anything, user).Return(([]UserOrderResponse)(nil), unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_user_orders", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("GET", "/user/orders", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetUserOrders(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderByID_UnknownError checks that unknown errors are handled properly.
func TestHandlerGetOrderByID_UnknownError(t *testing.T) {
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
	mockOrderService.On("GetOrderByID", mock.Anything, orderID, user).Return(nil, unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_order_by_id", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("GET", "/orders/"+orderID, nil)
	req = setChiURLParam(req, "orderID", orderID)

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderByID(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetOrderItemsByOrderID_UnknownError checks that unknown errors are handled properly.
func TestHandlerGetOrderItemsByOrderID_UnknownError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	orderID := testOrderID
	unknownError := errors.New("unknown database error")
	mockOrderService.On("GetOrderItemsByOrderID", mock.Anything, orderID).Return(([]OrderItemResponse)(nil), unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "get_order_items", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("GET", "/orders/"+orderID+"/items", nil)
	req = setChiURLParam(req, "order_id", orderID)

	w := httptest.NewRecorder()

	cfg.HandlerGetOrderItemsByOrderID(w, req, database.User{})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
