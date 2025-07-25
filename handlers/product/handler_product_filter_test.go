// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_product_filter_test.go: Tests the filter products handler for success, invalid input, and service error with expected responses and logging.

// TestHandlerFilterProducts_Success tests the successful filtering of products via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns products without error.
func TestHandlerFilterProducts_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1"}
	params := FilterProductsRequest{}
	products := []database.Product{{ID: "p1"}, {ID: "p2"}}
	mockService.On("FilterProducts", mock.Anything, params).Return(products, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "filter_products", "Filter products success", mock.Anything, mock.Anything).Return()
	jsonBody, _ := json.Marshal(params)

	req := httptest.NewRequest("POST", "/products/filter", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerFilterProducts_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerFilterProducts_InvalidPayload(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLog.On("LogHandlerError", mock.Anything, "filter_products", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/products/filter", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerFilterProducts_ServiceError tests the handler's behavior when the product service returns an error during filtering.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerFilterProducts_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1"}
	params := FilterProductsRequest{}
	err := &handlers.AppError{Code: "transaction_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("FilterProducts", mock.Anything, params).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "filter_products", "transaction_error", "fail", mock.Anything, mock.Anything, err.Err).Return()
	jsonBody, _ := json.Marshal(params)

	req := httptest.NewRequest("POST", "/products/filter", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
