package orderhandlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestInitOrderService_Success verifies successful initialization of the OrderService with all dependencies present.
func TestInitOrderService_Success(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: &sql.DB{},
			},
		},
	}

	err := cfg.InitOrderService()
	assert.NoError(t, err)
	assert.NotNil(t, cfg.orderService)
}

// TestInitOrderService_MissingHandlersConfig checks that initialization fails gracefully when HandlersConfig is missing.
func TestInitOrderService_MissingHandlersConfig(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: nil,
	}

	err := cfg.InitOrderService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

// TestInitOrderService_MissingDB checks that initialization fails gracefully when the database is missing.
func TestInitOrderService_MissingDB(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     nil,
				DBConn: &sql.DB{},
			},
		},
	}

	err := cfg.InitOrderService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestInitOrderService_MissingDBConn checks that initialization fails gracefully when the database connection is missing.
func TestInitOrderService_MissingDBConn(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: nil,
			},
		},
	}

	err := cfg.InitOrderService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// TestGetOrderService_Initialized verifies that GetOrderService returns the initialized service.
func TestGetOrderService_Initialized(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: &sql.DB{},
			},
		},
	}

	// Initialize the service first
	err := cfg.InitOrderService()
	if err == nil {
		// Test getting the service
		service := cfg.GetOrderService()
		assert.NotNil(t, service)
		assert.Equal(t, cfg.orderService, service)
	}
}

// TestGetOrderService_NotInitialized checks that GetOrderService auto-initializes the service if not already done.
func TestGetOrderService_NotInitialized(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: &sql.DB{},
			},
		},
	}

	// Test getting service without initialization (should auto-initialize)
	service := cfg.GetOrderService()
	assert.NotNil(t, service)
	assert.NotNil(t, cfg.orderService)
}

// TestGetOrderService_ThreadSafety checks that GetOrderService is thread-safe under concurrent access.
func TestGetOrderService_ThreadSafety(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: &sql.DB{},
			},
		},
	}

	// Test concurrent access to GetOrderService
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			service := cfg.GetOrderService()
			assert.NotNil(t, service)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify service was initialized
	assert.NotNil(t, cfg.orderService)
}

// TestGetOrderService_NilDependencies checks that GetOrderService handles nil dependencies gracefully.
func TestGetOrderService_NilDependencies(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: nil,
	}

	service := cfg.GetOrderService()
	assert.NotNil(t, service)

	// Test that service with nil deps doesn't panic
	_, err := service.GetAllOrders(context.Background())
	assert.Error(t, err)
}

// TestHandleOrderError_AllErrorCodes verifies handleOrderError returns correct status codes and messages for all error codes.
func TestHandleOrderError_AllErrorCodes(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}

	req := httptest.NewRequest("GET", "/test", nil)

	testCases := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "transaction error",
			err: &handlers.AppError{
				Code:    "transaction_error",
				Message: "Transaction failed",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later",
		},
		{
			name: "create order error",
			err: &handlers.AppError{
				Code:    "create_order_error",
				Message: "Failed to create order",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later",
		},
		{
			name: "order not found",
			err: &handlers.AppError{
				Code:    "order_not_found",
				Message: "Order not found",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Order not found",
		},
		{
			name: "invalid request",
			err: &handlers.AppError{
				Code:    "invalid_request",
				Message: "Invalid request",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request",
		},
		{
			name: "invalid status",
			err: &handlers.AppError{
				Code:    "invalid_status",
				Message: "Invalid status",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid status",
		},
		{
			name: "quantity overflow",
			err: &handlers.AppError{
				Code:    "quantity_overflow",
				Message: "Quantity overflow",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Quantity overflow",
		},
		{
			name: "unauthorized",
			err: &handlers.AppError{
				Code:    "unauthorized",
				Message: "Access denied",
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access denied",
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			// Set up mock expectations for logging
			mockLogger.On("LogHandlerError", mock.Anything, "test_operation", mock.Anything, mock.Anything, "127.0.0.1", "test-agent", mock.Anything).Return()

			cfg.handleOrderError(w, req, tc.err, "test_operation", "127.0.0.1", "test-agent")

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
		})
	}
}

// TestRequestResponseStructs tests the request and response structs for correct field assignment.
func TestRequestResponseStructs(t *testing.T) {
	// Test CreateOrderRequest
	createReq := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2, Price: 10.50},
		},
		PaymentMethod:     "credit_card",
		ShippingAddress:   "123 Main St",
		ContactPhone:      "555-1234",
		ExternalPaymentID: "ext_pay_123",
		TrackingNumber:    "TRK123",
	}
	assert.NotEmpty(t, createReq.Items)
	assert.Equal(t, "credit_card", createReq.PaymentMethod)
	assert.Equal(t, "123 Main St", createReq.ShippingAddress)
	assert.Equal(t, "555-1234", createReq.ContactPhone)
	assert.Equal(t, "ext_pay_123", createReq.ExternalPaymentID)
	assert.Equal(t, "TRK123", createReq.TrackingNumber)

	// Test OrderItemInput
	itemInput := OrderItemInput{
		ProductID: "prod1",
		Quantity:  5,
		Price:     25.99,
	}
	assert.Equal(t, "prod1", itemInput.ProductID)
	assert.Equal(t, 5, itemInput.Quantity)
	assert.Equal(t, 25.99, itemInput.Price)

	// Test UpdateOrderStatusRequest
	updateReq := UpdateOrderStatusRequest{
		Status: "shipped",
	}
	assert.Equal(t, "shipped", updateReq.Status)

	// Test OrderItemResponse
	itemResp := OrderItemResponse{
		ID:        "item1",
		ProductID: "prod1",
		Quantity:  3,
		Price:     "15.99",
	}
	assert.Equal(t, "item1", itemResp.ID)
	assert.Equal(t, "prod1", itemResp.ProductID)
	assert.Equal(t, 3, itemResp.Quantity)
	assert.Equal(t, "15.99", itemResp.Price)

	// Test UserOrderResponse
	userOrderResp := UserOrderResponse{
		OrderID:         "order1",
		TotalAmount:     "47.97",
		Status:          "pending",
		PaymentMethod:   "credit_card",
		TrackingNumber:  "TRK123",
		ShippingAddress: "123 Main St",
		ContactPhone:    "555-1234",
		Items:           []OrderItemResponse{itemResp},
	}
	assert.Equal(t, "order1", userOrderResp.OrderID)
	assert.Equal(t, "47.97", userOrderResp.TotalAmount)
	assert.Equal(t, "pending", userOrderResp.Status)
	assert.Equal(t, "credit_card", userOrderResp.PaymentMethod)
	assert.Equal(t, "TRK123", userOrderResp.TrackingNumber)
	assert.Equal(t, "123 Main St", userOrderResp.ShippingAddress)
	assert.Equal(t, "555-1234", userOrderResp.ContactPhone)
	assert.Len(t, userOrderResp.Items, 1)

	// Test OrderResponse
	orderResp := OrderResponse{
		Message: "Order created successfully",
		OrderID: "order1",
	}
	assert.Equal(t, "Order created successfully", orderResp.Message)
	assert.Equal(t, "order1", orderResp.OrderID)
}

// TestHandlersOrderConfig_ConcurrentAccess tests that the HandlersOrderConfig can handle concurrent access safely.
func TestHandlersOrderConfig_ConcurrentAccess(t *testing.T) {
	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: &sql.DB{},
			},
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent initialization
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cfg.InitOrderService()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// Test concurrent service access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service := cfg.GetOrderService()
			assert.NotNil(t, service)
		}()
	}
	wg.Wait()
}
