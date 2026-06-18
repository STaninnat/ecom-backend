// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
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
	params.CategoryID.Valid = true
	params.CategoryID.String = "cat1"
	params.IsActive.Valid = true
	params.IsActive.Bool = true
	params.MinPrice.Valid = true
	params.MinPrice.Float64 = 10.5
	params.MaxPrice.Valid = true
	params.MaxPrice.Float64 = 100.0

	products := []database.Product{{ID: "00000000-0000-0000-0000-000000000001"}, {ID: "p2"}}
	mockService.On("FilterProducts", mock.Anything, params).Return(products, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "filter_products", "Filter products success", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/products/filter?category_id=cat1&is_active=true&min_price=10.5&max_price=100", nil)
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

func TestHandlerFilterProducts_InvalidMinPrice(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1"}
	mockLog.On("LogHandlerError", mock.Anything, "filter_products", "invalid_request", "Invalid min_price", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/products/filter?min_price=abc", nil)
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

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

	req := httptest.NewRequest("GET", "/products/filter", nil)
	w := httptest.NewRecorder()

	cfg.HandlerFilterProducts(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
