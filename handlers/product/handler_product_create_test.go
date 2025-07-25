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
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_product_create_test.go: Tests the product creation handler for success, invalid input, and service error scenarios with proper responses and logging.

// TestHandlerCreateProduct_Success tests the successful creation of a product via the handler.
// It verifies that the handler returns HTTP 201, the correct response message, and product ID when the service succeeds.
func TestHandlerCreateProduct_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	params := ProductRequest{CategoryID: "c1", Name: "P", Price: 10, Stock: 1}
	jsonBody, _ := json.Marshal(params)
	mockService.On("CreateProduct", mock.Anything, params).Return("pid1", nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "create_product", "Created product successful", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateProduct(w, req, user)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp productResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Product created successfully", resp.Message)
	assert.Equal(t, "pid1", resp.ProductID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerProduct_InvalidPayload tests invalid JSON payloads for both create and update product handlers.
func TestHandlerProduct_InvalidPayload(t *testing.T) {
	tests := []struct {
		name        string
		handlerFunc func(cfg *HandlersProductConfig, w http.ResponseWriter, req *http.Request, user database.User)
		logOp       string
		method      string
		url         string
	}{
		{
			name: "CreateProduct_InvalidPayload",
			handlerFunc: func(cfg *HandlersProductConfig, w http.ResponseWriter, req *http.Request, user database.User) {
				cfg.HandlerCreateProduct(w, req, user)
			},
			logOp:  "create_product",
			method: "POST",
			url:    "/products",
		},
		{
			name: "UpdateProduct_InvalidPayload",
			handlerFunc: func(cfg *HandlersProductConfig, w http.ResponseWriter, req *http.Request, user database.User) {
				cfg.HandlerUpdateProduct(w, req, user)
			},
			logOp:  "update_product",
			method: "PUT",
			url:    "/products/pid1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			mockLog.On("LogHandlerError", mock.Anything, tc.logOp, "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

			req := httptest.NewRequest(tc.method, tc.url, bytes.NewBuffer(badBody))
			w := httptest.NewRecorder()

			tc.handlerFunc(cfg, w, req, user)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			mockLog.AssertExpectations(t)
		})
	}
}

// TestHandlerCreateProduct_ServiceError tests the handler's behavior when the product service returns an error.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerCreateProduct_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1"}
	params := ProductRequest{CategoryID: "c1", Name: "P", Price: 10, Stock: 1}
	jsonBody, _ := json.Marshal(params)
	err := &handlers.AppError{Code: "create_product_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("CreateProduct", mock.Anything, params).Return("", err)
	mockLog.On("LogHandlerError", mock.Anything, "create_product", "create_product_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateProduct(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
