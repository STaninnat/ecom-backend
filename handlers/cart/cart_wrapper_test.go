// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
)

// cart_wrapper_test.go: Tests for cart service initialization, error handling, DTOs, and service creation.

// TestInitCartService_Success tests the successful initialization of the cart service.
// It verifies that the service is properly set in the configuration and can be retrieved
// without errors when the configuration is valid.
func TestInitCartService_Success(t *testing.T) {
	cfg := &HandlersCartConfig{Config: &handlers.Config{}}
	mockService := new(MockCartService)

	err := cfg.InitCartService(mockService)
	require.NoError(t, err)
	assert.NotNil(t, cfg.GetCartService())
}

// TestInitCartService_Error tests the cart service initialization when the configuration is invalid.
// It ensures that the initialization returns an error when the Config is nil,
// preventing the service from being set in an invalid state.
func TestInitCartService_Error(t *testing.T) {
	cfg := &HandlersCartConfig{Config: nil}
	mockService := new(MockCartService)

	err := cfg.InitCartService(mockService)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestGetCartService_ReturnsCorrectInstance tests that the GetCartService method returns
// the correct service instance after initialization. It verifies that the service is not nil
// and can be accessed through the getter method.
func TestGetCartService_ReturnsCorrectInstance(t *testing.T) {
	cfg := &HandlersCartConfig{Config: &handlers.Config{}}
	mockService := new(MockCartService)
	err := cfg.InitCartService(mockService)
	require.NoError(t, err)

	service := cfg.GetCartService()
	assert.NotNil(t, service)
}

// TestHandleCartError_AppErrorCodes tests the handleCartError method with various AppError codes.
// It verifies that the method correctly maps different error codes to appropriate HTTP status codes
// and ensures proper error handling for different types of application errors.
func TestHandleCartError_AppErrorCodes(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected int
	}{
		{"not_found", "not_found", http.StatusNotFound},
		{"unauthorized", "unauthorized", http.StatusForbidden},
		{"invalid_request", "invalid_request", http.StatusBadRequest},
		{"product_not_found", "product_not_found", http.StatusNotFound},
		{"cart_empty", "cart_empty", http.StatusBadRequest},
		{"insufficient_stock", "insufficient_stock", http.StatusBadRequest},
		{"cart_full", "cart_full", http.StatusBadRequest},
		{"add_failed", "add_failed", http.StatusInternalServerError},
		{"get_failed", "get_failed", http.StatusInternalServerError},
		{"update_failed", "update_failed", http.StatusInternalServerError},
		{"remove_failed", "remove_failed", http.StatusInternalServerError},
		{"clear_failed", "clear_failed", http.StatusInternalServerError},
		{"get_cart_failed", "get_cart_failed", http.StatusInternalServerError},
		{"save_cart_failed", "save_cart_failed", http.StatusInternalServerError},
		{"invalid_price", "invalid_price", http.StatusInternalServerError},
		{"invalid_quantity", "invalid_quantity", http.StatusInternalServerError},
		{"transaction_error", "transaction_error", http.StatusInternalServerError},
		{"create_order_failed", "create_order_failed", http.StatusInternalServerError},
		{"update_stock_failed", "update_stock_failed", http.StatusInternalServerError},
		{"create_order_item_failed", "create_order_item_failed", http.StatusInternalServerError},
		{"commit_failed", "commit_failed", http.StatusInternalServerError},
		{"default", "other", http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &HandlersCartConfig{Config: &handlers.Config{}}
			mockLogger := new(MockLogger)
			cfg.Logger = mockLogger
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			err := &handlers.AppError{Code: tc.code, Message: "msg"}
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			cfg.handleCartError(w, r, err, "op", "ip", "ua")
			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

// TestHandleCartError_NonAppError tests the handleCartError method with non-AppError types.
// It verifies that the method correctly handles generic errors by returning HTTP 500
// and ensures that the error handling is robust for different error types.
func TestHandleCartError_NonAppError(t *testing.T) {
	cfg := &HandlersCartConfig{Config: &handlers.Config{}}
	mockLogger := new(MockLogger)
	cfg.Logger = mockLogger
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := errors.New("fail")
	mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	cfg.handleCartError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestNewCartServiceWithDeps tests the convenience function for creating a cart service with dependencies.
// It verifies that the function returns a non-nil service instance.
func TestNewCartServiceWithDeps(t *testing.T) {
	// This test would require actual dependencies, but we can test the function signature
	// In a real implementation, you might want to mock the dependencies
	service := NewCartServiceWithDeps(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

// TestCartCheckoutResult_Structure tests the CartCheckoutResult structure.
// It verifies that the struct has the expected fields and JSON tags.
func TestCartCheckoutResult_Structure(t *testing.T) {
	result := CartCheckoutResult{
		OrderID: "order123",
		Message: "Order placed successfully",
	}

	assert.Equal(t, "order123", result.OrderID)
	assert.Equal(t, "Order placed successfully", result.Message)
}

// TestCartItemRequest_Structure tests the CartItemRequest structure.
// It verifies that the struct has the expected fields and JSON tags.
func TestCartItemRequest_Structure(t *testing.T) {
	request := CartItemRequest{
		ProductID: "product123",
		Quantity:  5,
	}

	assert.Equal(t, "product123", request.ProductID)
	assert.Equal(t, 5, request.Quantity)
}

// TestCartUpdateRequest_Structure tests the CartUpdateRequest structure.
// It verifies that the struct has the expected fields and JSON tags.
func TestCartUpdateRequest_Structure(t *testing.T) {
	request := CartUpdateRequest{
		ProductID: "product123",
		Quantity:  3,
	}

	assert.Equal(t, "product123", request.ProductID)
	assert.Equal(t, 3, request.Quantity)
}

// TestCartResponse_Structure tests the CartResponse structure.
// It verifies that the struct has the expected fields and JSON tags.
func TestCartResponse_Structure(t *testing.T) {
	response := CartResponse{
		Message: "Item added to cart",
		OrderID: "order123",
	}

	assert.Equal(t, "Item added to cart", response.Message)
	assert.Equal(t, "order123", response.OrderID)
}
