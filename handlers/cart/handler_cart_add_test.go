// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_cart_add_test.go: Tests for AddItemToUserCart and AddItemToGuestCart handlers, validating request handling and error scenarios.

const (
	testSessionIDAdd = "sess1"
)

// TestHandlerAddItemToUserCart_Success tests adding an item to a user's cart successfully.
// It verifies that the handler returns HTTP 200 and calls the service and logger as expected.
func TestHandlerAddItemToUserCart_Success(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	user := database.User{ID: "u1"}
	reqBody := CartItemRequest{ProductID: "p1", Quantity: 2}
	jsonBody, _ := json.Marshal(reqBody)
	mockService.On("AddItemToUserCart", mock.Anything, user.ID, reqBody.ProductID, reqBody.Quantity).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "add_item_to_cart", "Added item to cart", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerAddItemToUserCart(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToUserCart_InvalidJSON tests the handler with invalid JSON in the request body.
// It verifies that the handler returns HTTP 400 and logs the error.
func TestHandlerAddItemToUserCart_InvalidJSON(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	user := database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_to_cart", "invalid request body", "Failed to parse body", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerAddItemToUserCart(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToUserCart_MissingFields tests the handler with missing required fields in the request body.
// It verifies that the handler returns HTTP 400 and logs the error.
func TestHandlerAddItemToUserCart_MissingFields(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	user := database.User{ID: "u1"}
	reqBody := CartItemRequest{ProductID: "", Quantity: 0}
	jsonBody, _ := json.Marshal(reqBody)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_to_cart", "missing fields", "Required fields are missing", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerAddItemToUserCart(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToUserCart_ServiceError tests the handler when the service returns an application error.
// It verifies that the handler returns the appropriate HTTP status and logs the error.
func TestHandlerAddItemToUserCart_ServiceError(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	user := database.User{ID: "u1"}
	reqBody := CartItemRequest{ProductID: "p1", Quantity: 2}
	jsonBody, _ := json.Marshal(reqBody)
	err := &handlers.AppError{Code: "product_not_found", Message: "not found", Err: errors.New("fail")}
	mockService.On("AddItemToUserCart", mock.Anything, user.ID, reqBody.ProductID, reqBody.Quantity).Return(err)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_to_cart", "product_not_found", "not found", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerAddItemToUserCart(w, req, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToGuestCart_Success tests adding an item to a guest cart successfully.
// It verifies that the handler returns HTTP 200 and calls the service and logger as expected.
func TestHandlerAddItemToGuestCart_Success(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	reqBody := CartItemRequest{ProductID: "p1", Quantity: 2}
	jsonBody, _ := json.Marshal(reqBody)
	mockService.On("AddItemToGuestCart", mock.Anything, testSessionIDAdd, reqBody.ProductID, reqBody.Quantity).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "add_item_guest_cart", "Added item to guest cart", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.WithValue(req.Context(), utils.ContextKey("session_id"), testSessionIDAdd))
	w := httptest.NewRecorder()

	// Patch getSessionIDFromRequest to return "sess1"
	orig := getSessionIDFromRequest
	getSessionIDFromRequest = func(_ *http.Request) string { return testSessionIDAdd }
	defer func() { getSessionIDFromRequest = orig }()

	cfg.HandlerAddItemToGuestCart(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToGuestCart_MissingSessionID tests the handler when the session ID is missing from the request.
// It verifies that the handler returns HTTP 400 and logs the error.
func TestHandlerAddItemToGuestCart_MissingSessionID(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_guest_cart", "missing session ID", "Session ID not found in request", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", nil)
	w := httptest.NewRecorder()

	// Patch getSessionIDFromRequest to return ""
	orig := getSessionIDFromRequest
	getSessionIDFromRequest = func(_ *http.Request) string { return "" }
	defer func() { getSessionIDFromRequest = orig }()

	cfg.HandlerAddItemToGuestCart(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToGuestCart_InvalidJSON tests the handler with invalid JSON in the request body for a guest cart.
// It verifies that the handler returns HTTP 400 and logs the error.
func TestHandlerAddItemToGuestCart_InvalidJSON(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	badBody := []byte(`{"bad":}`)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_guest_cart", "invalid request body", "Failed to parse body", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	// Patch getSessionIDFromRequest to return "sess1"
	orig := getSessionIDFromRequest
	getSessionIDFromRequest = func(_ *http.Request) string { return testSessionIDAdd }
	defer func() { getSessionIDFromRequest = orig }()

	cfg.HandlerAddItemToGuestCart(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToGuestCart_MissingFields tests the handler with missing required fields in the request body for a guest cart.
// It verifies that the handler returns HTTP 400 and logs the error.
func TestHandlerAddItemToGuestCart_MissingFields(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	reqBody := CartItemRequest{ProductID: "", Quantity: 0}
	jsonBody, _ := json.Marshal(reqBody)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_guest_cart", "missing fields", "Required fields are missing", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	// Patch getSessionIDFromRequest to return "sess1"
	orig := getSessionIDFromRequest
	getSessionIDFromRequest = func(_ *http.Request) string { return testSessionIDAdd }
	defer func() { getSessionIDFromRequest = orig }()

	cfg.HandlerAddItemToGuestCart(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerAddItemToGuestCart_ServiceError tests the handler when the service returns an application error for a guest cart.
// It verifies that the handler returns the appropriate HTTP status and logs the error.
func TestHandlerAddItemToGuestCart_ServiceError(t *testing.T) {
	mockService := new(MockCartService)
	mockLogger := new(MockLogger)
	cfg := &HandlersCartConfig{
		Config:      &handlers.Config{},
		Logger:      mockLogger,
		CartService: mockService,
	}
	reqBody := CartItemRequest{ProductID: "p1", Quantity: 2}
	jsonBody, _ := json.Marshal(reqBody)
	err := &handlers.AppError{Code: "cart_full", Message: "full", Err: errors.New("fail")}
	mockService.On("AddItemToGuestCart", mock.Anything, testSessionIDAdd, reqBody.ProductID, reqBody.Quantity).Return(err)
	mockLogger.On("LogHandlerError", mock.Anything, "add_item_guest_cart", "cart_full", "full", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("POST", "/cart", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	// Patch getSessionIDFromRequest to return "sess1"
	orig := getSessionIDFromRequest
	getSessionIDFromRequest = func(_ *http.Request) string { return testSessionIDAdd }
	defer func() { getSessionIDFromRequest = orig }()

	cfg.HandlerAddItemToGuestCart(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
