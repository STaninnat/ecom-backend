// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/stretchr/testify/assert"
)

// upload_wrapper_test.go: Tests all error handling branches of handleUploadError for local and S3 upload configs,
// verifying correct HTTP responses and proper logging in each case.

// TestHandlersUploadConfig_handleUploadError_AllBranches tests handleUploadError for HandlersUploadConfig.
// It verifies that all error branches are handled correctly, the right status code and message are returned, and logging is performed as expected.
func TestHandlersUploadConfig_handleUploadError_AllBranches(t *testing.T) {
	mockLogger := new(MockLogger)
	cfg := &HandlersUploadConfig{Logger: mockLogger}
	req := httptest.NewRequest("POST", "/upload", nil)
	ctx := req.Context()
	ip, userAgent := "1.2.3.4", "test-agent"

	testCases := []struct {
		name       string
		err        error
		expectCode int
		expectMsg  string
		logCode    string
		logErr     error
	}{
		{"missing_product_id", &handlers.AppError{Code: "missing_product_id", Message: "Missing product"}, http.StatusBadRequest, "Missing product", "missing_product_id", nil},
		{"not_found", &handlers.AppError{Code: "not_found", Message: "Not found", Err: errors.New("db")}, http.StatusNotFound, "Not found", "not_found", errors.New("db")},
		{"invalid_form", &handlers.AppError{Code: "invalid_form", Message: "Invalid form", Err: errors.New("form")}, http.StatusBadRequest, "Invalid form", "invalid_form", errors.New("form")},
		{"invalid_image", &handlers.AppError{Code: "invalid_image", Message: "Invalid image", Err: errors.New("img")}, http.StatusBadRequest, "Invalid image", "invalid_image", errors.New("img")},
		{"db_error", &handlers.AppError{Code: "db_error", Message: "DB fail", Err: errors.New("db")}, http.StatusInternalServerError, "Something went wrong, please try again later", "db_error", errors.New("db")},
		{"file_save_failed", &handlers.AppError{Code: "file_save_failed", Message: "Save fail", Err: errors.New("fs")}, http.StatusInternalServerError, "Something went wrong, please try again later", "file_save_failed", errors.New("fs")},
		{"transaction_error", &handlers.AppError{Code: "transaction_error", Message: "Tx fail", Err: errors.New("tx")}, http.StatusInternalServerError, "Something went wrong, please try again later", "transaction_error", errors.New("tx")},
		{"commit_error", &handlers.AppError{Code: "commit_error", Message: "Commit fail", Err: errors.New("commit")}, http.StatusInternalServerError, "Something went wrong, please try again later", "commit_error", errors.New("commit")},
		{"default_app_error", &handlers.AppError{Code: "other", Message: "Other fail", Err: errors.New("other")}, http.StatusInternalServerError, "Internal server error", "internal_error", errors.New("other")},
		{"generic_error", errors.New("something bad"), http.StatusInternalServerError, "Internal server error", "unknown_error", errors.New("something bad")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{}
			if errors.As(tc.err, &appErr) {
				mockLogger.On("LogHandlerError", ctx, "op", tc.logCode, appErr.Message, ip, userAgent, tc.logErr).Return().Once()
			} else {
				mockLogger.On("LogHandlerError", ctx, "op", tc.logCode, "Unknown error occurred", ip, userAgent, tc.logErr).Return().Once()
			}
			cfg.handleUploadError(w, req, tc.err, "op", ip, userAgent)
			assert.Equal(t, tc.expectCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectMsg)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlersUploadS3Config_handleUploadError_AllBranches tests handleUploadError for HandlersUploadS3Config.
// It verifies that all error branches are handled correctly, the right status code and message are returned, and logging is performed as expected.
func TestHandlersUploadS3Config_handleUploadError_AllBranches(t *testing.T) {
	mockLogger := new(MockLogger)
	cfg := &HandlersUploadS3Config{Logger: mockLogger}
	req := httptest.NewRequest("POST", "/upload", nil)
	ctx := req.Context()
	ip, userAgent := "1.2.3.4", "test-agent"

	testCases := []struct {
		name       string
		err        error
		expectCode int
		expectMsg  string
		logCode    string
		logErr     error
	}{
		{"missing_product_id", &handlers.AppError{Code: "missing_product_id", Message: "Missing product"}, http.StatusBadRequest, "Missing product", "missing_product_id", nil},
		{"not_found", &handlers.AppError{Code: "not_found", Message: "Not found", Err: errors.New("db")}, http.StatusNotFound, "Not found", "not_found", errors.New("db")},
		{"invalid_form", &handlers.AppError{Code: "invalid_form", Message: "Invalid form", Err: errors.New("form")}, http.StatusBadRequest, "Invalid form", "invalid_form", errors.New("form")},
		{"invalid_image", &handlers.AppError{Code: "invalid_image", Message: "Invalid image", Err: errors.New("img")}, http.StatusBadRequest, "Invalid image", "invalid_image", errors.New("img")},
		{"db_error", &handlers.AppError{Code: "db_error", Message: "DB fail", Err: errors.New("db")}, http.StatusInternalServerError, "Something went wrong, please try again later", "db_error", errors.New("db")},
		{"file_save_failed", &handlers.AppError{Code: "file_save_failed", Message: "Save fail", Err: errors.New("fs")}, http.StatusInternalServerError, "Something went wrong, please try again later", "file_save_failed", errors.New("fs")},
		{"transaction_error", &handlers.AppError{Code: "transaction_error", Message: "Tx fail", Err: errors.New("tx")}, http.StatusInternalServerError, "Something went wrong, please try again later", "transaction_error", errors.New("tx")},
		{"commit_error", &handlers.AppError{Code: "commit_error", Message: "Commit fail", Err: errors.New("commit")}, http.StatusInternalServerError, "Something went wrong, please try again later", "commit_error", errors.New("commit")},
		{"default_app_error", &handlers.AppError{Code: "other", Message: "Other fail", Err: errors.New("other")}, http.StatusInternalServerError, "Internal server error", "internal_error", errors.New("other")},
		{"generic_error", errors.New("something bad"), http.StatusInternalServerError, "Internal server error", "unknown_error", errors.New("something bad")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{}
			if errors.As(tc.err, &appErr) {
				mockLogger.On("LogHandlerError", ctx, "op", tc.logCode, appErr.Message, ip, userAgent, tc.logErr).Return().Once()
			} else {
				mockLogger.On("LogHandlerError", ctx, "op", tc.logCode, "Unknown error occurred", ip, userAgent, tc.logErr).Return().Once()
			}
			cfg.handleUploadError(w, req, tc.err, "op", ip, userAgent)
			assert.Equal(t, tc.expectCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectMsg)
			mockLogger.AssertExpectations(t)
		})
	}
}
