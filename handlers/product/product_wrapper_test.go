// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"database/sql"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// product_wrapper_test.go: Tests for product service initialization, lazy loading, and error handling with logging and HTTP responses.

// TestInitProductService_MissingConfig tests InitProductService when both DB and DBConn are nil.
// It expects an error indicating the database is not initialized.
func TestInitProductService_MissingConfig(t *testing.T) {
	cfg := &HandlersProductConfig{DB: nil, DBConn: nil, Logger: nil}
	err := cfg.InitProductService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestInitProductService_MissingDB tests InitProductService when DB is nil.
// It expects an error indicating the database is not initialized.
func TestInitProductService_MissingDB(t *testing.T) {
	cfg := &HandlersProductConfig{DB: nil, DBConn: nil, Logger: nil}
	err := cfg.InitProductService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestInitProductService_MissingDBConn tests InitProductService when DBConn is nil but DB is set.
// It expects an error indicating the database connection is not initialized.
func TestInitProductService_MissingDBConn(t *testing.T) {
	cfg := &HandlersProductConfig{DB: new(database.Queries), DBConn: nil, Logger: nil}
	err := cfg.InitProductService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// TestGetProductService_LazyInit tests that GetProductService lazily initializes the product service if not already set.
// It expects a non-nil service to be returned.
func TestGetProductService_LazyInit(t *testing.T) {
	cfg := &HandlersProductConfig{DB: nil, DBConn: nil, Logger: nil}
	svc := cfg.GetProductService()
	assert.NotNil(t, svc)
}

// TestGetProductService_AlreadySet tests that GetProductService returns the already set productService instance.
func TestGetProductService_AlreadySet(t *testing.T) {
	cfg := &HandlersProductConfig{DB: new(database.Queries), DBConn: new(sql.DB), Logger: nil}
	mockSvc := new(MockProductService)
	cfg.productService = mockSvc
	result := cfg.GetProductService()
	assert.Equal(t, mockSvc, result)
}

// TestGetProductService_OneNil tests GetProductService when only one of DB or DBConn is set.
// It expects a non-nil service to be returned in both cases.
func TestGetProductService_OneNil(t *testing.T) {
	cfg := &HandlersProductConfig{DB: new(database.Queries), DBConn: nil, Logger: nil}
	result := cfg.GetProductService()
	assert.NotNil(t, result)
	cfg2 := &HandlersProductConfig{DB: nil, DBConn: new(sql.DB), Logger: nil}
	result2 := cfg2.GetProductService()
	assert.NotNil(t, result2)
}

// TestHandleProductError_KnownError tests handleProductError with a known AppError code.
// It expects the correct HTTP status and logging behavior.
func TestHandleProductError_KnownError(t *testing.T) {
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{Logger: mockLog}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	err := &handlers.AppError{Code: "update_failed", Message: "fail", Err: errors.New("db")}
	mockLog.On("LogHandlerError", mock.Anything, "op", "update_failed", "fail", "ip", "ua", err.Err).Return()
	cfg.handleProductError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandleProductError_AllCodes tests handleProductError for all known AppError codes and a non-AppError.
// It expects the correct HTTP status for each code and verifies logging.
func TestHandleProductError_AllCodes(t *testing.T) {
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{Logger: mockLog}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	testCases := []struct {
		code     string
		wantCode int
	}{
		{"transaction_error", http.StatusInternalServerError},
		{"update_failed", http.StatusInternalServerError},
		{"commit_error", http.StatusInternalServerError},
		{"create_product_error", http.StatusInternalServerError},
		{"delete_product_error", http.StatusInternalServerError},
		{"product_not_found", http.StatusNotFound},
		{"invalid_request", http.StatusBadRequest},
		{"unknown_code", http.StatusInternalServerError},
	}
	for _, tc := range testCases {
		err := &handlers.AppError{Code: tc.code, Message: "fail", Err: errors.New("fail")}
		mockLog.On("LogHandlerError", mock.Anything, "op", mock.Anything, mock.Anything, "ip", "ua", err.Err).Return()
		cfg.handleProductError(w, r, err, "op", "ip", "ua")
		assert.Equal(t, tc.wantCode, w.Code)
		w = httptest.NewRecorder() // reset for next case
	}
	// Non-AppError
	err := errors.New("fail")
	mockLog.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()
	cfg.handleProductError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
