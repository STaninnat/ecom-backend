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
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Product created successfully", resp.Message)
	assert.Equal(t, "pid1", resp.ProductID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerCreateProduct_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerCreateProduct_InvalidPayload(t *testing.T) {
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
	mockLog.On("LogHandlerError", mock.Anything, "create_product", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateProduct(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
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
