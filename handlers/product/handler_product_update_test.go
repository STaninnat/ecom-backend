// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_product_update_test.go: Tests the update product handler for success, invalid input, and service error with expected responses and logging.

// TestHandlerUpdateProduct_Success tests the successful update of a product via the handler.
// It verifies that the handler returns HTTP 200, the correct response message, and logs success when the service updates the product without error.
func TestHandlerUpdateProduct_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	params := ProductRequest{ID: "pid1", CategoryID: "c1", Name: "P", Price: 10, Stock: 1}
	jsonBody, _ := json.Marshal(params)
	mockService.On("UpdateProduct", mock.Anything, params).Return(nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "update_product", "Updated product successfully", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("PUT", "/products/pid1", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateProduct(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.HandlerResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "Product updated successfully", resp.Message)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerUpdateProduct_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerUpdateProduct_InvalidPayload(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLog.On("LogHandlerError", mock.Anything, "update_product", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("PUT", "/products/pid1", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateProduct(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerUpdateProduct_ServiceError tests the handler's behavior when the product service returns an error during update.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerUpdateProduct_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	params := ProductRequest{ID: "pid1", CategoryID: "c1", Name: "P", Price: 10, Stock: 1}
	jsonBody, _ := json.Marshal(params)
	err := &handlers.AppError{Code: "update_failed", Message: "fail", Err: errors.New("fail")}
	mockService.On("UpdateProduct", mock.Anything, params).Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "update_product", "update_failed", "fail", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("PUT", "/products/pid1", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateProduct(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
