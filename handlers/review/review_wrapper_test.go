// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// review_wrapper_test.go: Tests for review handler config, service initialization, and error handling behavior.

// TestInitReviewService_Success tests the successful initialization of the review service.
// It verifies that the service is properly set in the configuration and can be retrieved
// without errors when the configuration is valid.
func TestInitReviewService_Success(t *testing.T) {
	cfg := &HandlersReviewConfig{Config: &handlers.Config{}}
	err := cfg.InitReviewService(&mockReviewService{})
	assert.NoError(t, err)
	assert.NotNil(t, cfg.GetReviewService())
}

// TestInitReviewService_Error tests the review service initialization when the configuration is invalid.
// It ensures that the initialization returns an error when the Config is nil,
// preventing the service from being set in an invalid state.
func TestInitReviewService_Error(t *testing.T) {
	cfg := &HandlersReviewConfig{Config: nil}
	err := cfg.InitReviewService(&mockReviewService{})
	assert.Error(t, err)
}

// TestGetReviewService_ReturnsCorrectInstance tests that the GetReviewService method returns
// the correct service instance after initialization. It verifies that the service is not nil
// and can be accessed through the getter method.
func TestGetReviewService_ReturnsCorrectInstance(t *testing.T) {
	cfg := &HandlersReviewConfig{Config: &handlers.Config{}}
	err := cfg.InitReviewService(&mockReviewService{})
	assert.NoError(t, err)
	rs := cfg.GetReviewService()
	assert.NotNil(t, rs)
}

// TestHandleReviewError_AppErrorCodes tests the handleReviewError method with various AppError codes.
// It verifies that the method correctly maps different error codes to appropriate HTTP status codes
// and ensures proper error handling for different types of application errors.
func TestHandleReviewError_AppErrorCodes(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected int
	}{
		{"not_found", "not_found", http.StatusNotFound},
		{"unauthorized", "unauthorized", http.StatusForbidden},
		{"invalid_request", "invalid_request", http.StatusBadRequest},
		{"default", "other", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &HandlersReviewConfig{Config: &handlers.Config{}}
			mockLogger := new(mockLoggerWrapper)
			cfg.Logger = mockLogger
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			err := &handlers.AppError{Code: tc.code, Message: "msg"}
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			cfg.handleReviewError(w, r, err, "op", "ip", "ua")
			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

// TestHandleReviewError_NonAppError tests the handleReviewError method with non-AppError types.
// It verifies that the method correctly handles generic errors by returning HTTP 500
// and ensures that the error handling is robust for different error types.
func TestHandleReviewError_NonAppError(t *testing.T) {
	cfg := &HandlersReviewConfig{Config: &handlers.Config{}}
	mockLogger := new(mockLoggerWrapper)
	cfg.Logger = mockLogger
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := errors.New("fail")
	mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	cfg.handleReviewError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
