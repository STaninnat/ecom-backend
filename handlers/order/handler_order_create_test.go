// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"bytes"
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

// handler_order_create_test.go: Tests for HandlerCreateOrder covering success, validation errors, service errors, and full request scenarios.

// TestHandlerCreateOrder_Success verifies successful order creation with valid input.
func TestHandlerCreateOrder_Success(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2, Price: 10.50},
			{ProductID: "prod2", Quantity: 1, Price: 25.00},
		},
		PaymentMethod:   "credit_card",
		ShippingAddress: "123 Main St",
		ContactPhone:    "555-1234",
	}

	expectedResult := &OrderResponse{
		Message: "Created order successful",
		OrderID: "order123",
	}

	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(expectedResult, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "create_order", "Created order successful", mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Created order successful", response.Message)
	assert.Equal(t, "order123", response.OrderID)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerCreateOrder_InvalidRequest checks that an invalid JSON request returns a bad request error.
func TestHandlerCreateOrder_InvalidRequest(t *testing.T) {
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
	invalidJSON := `{"items": [{"product_id": "prod1", "quantity": 2, "price": 10.50}`
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "invalid_request", "Failed to parse request body", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "CreateOrder")
}

// TestHandlerCreateOrder_EmptyItems ensures empty items list results in an error.
func TestHandlerCreateOrder_EmptyItems(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{},
	}

	appError := &handlers.AppError{Code: "invalid_request", Message: "Order must contain at least one item"}
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "invalid_request", "Order must contain at least one item", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order must contain at least one item", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_InvalidQuantity checks that invalid quantity returns the correct error.
func TestHandlerCreateOrder_InvalidQuantity(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 0, Price: 10.50},
		},
	}

	appError := &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "invalid_request", "Quantity must be greater than 0", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Quantity must be greater than 0", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_QuantityOverflow checks that quantity overflow returns the correct error.
func TestHandlerCreateOrder_QuantityOverflow(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2147483648, Price: 10.50}, // Exceeds int32 max
		},
	}

	appError := &handlers.AppError{Code: "quantity_overflow", Message: "Quantity 2147483648 exceeds the max limit for int32"}
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "quantity_overflow", "Quantity 2147483648 exceeds the max limit for int32", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Quantity 2147483648 exceeds the max limit for int32", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_TransactionError ensures a transaction error is handled correctly.
func TestHandlerCreateOrder_TransactionError(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 1, Price: 10.50},
		},
	}

	appError := &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction"}
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "transaction_error", "Error starting transaction", mock.Anything, mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_CreateOrderError ensures a create order error is handled correctly.
func TestHandlerCreateOrder_CreateOrderError(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 1, Price: 10.50},
		},
	}

	appError := &handlers.AppError{Code: "create_order_error", Message: "Error creating order"}
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, appError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "create_order_error", "Error creating order", mock.Anything, mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_UnknownError ensures an unknown error is handled correctly.
func TestHandlerCreateOrder_UnknownError(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 1, Price: 10.50},
		},
	}

	unknownError := errors.New("unknown database error")
	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestHandlerCreateOrder_CompleteRequest tests order creation with all optional fields.
func TestHandlerCreateOrder_CompleteRequest(t *testing.T) {
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
	requestBody := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2, Price: 10.50},
		},
		PaymentMethod:     "credit_card",
		ShippingAddress:   "123 Main St",
		ContactPhone:      "555-1234",
		ExternalPaymentID: "ext_pay_123",
		TrackingNumber:    "TRK123",
	}

	expectedResult := &OrderResponse{
		Message: "Created order successful",
		OrderID: "order123",
	}

	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(expectedResult, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "create_order", "Created order successful", mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Created order successful", response.Message)
	assert.Equal(t, "order123", response.OrderID)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
