// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_s3_test.go: Tests S3 product image upload and update handlers for success, error, and missing ID cases, verifying correct HTTP responses and proper logging behavior.

// TestHandlerS3UploadProductImage_Success tests the successful upload of a product image to S3 via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns an S3 image URL without error.
func TestHandlerS3UploadProductImage_Success(t *testing.T) {
	mockLogger := new(mockS3Logger)
	mockService := new(mockS3UploadService)
	cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()

	mockService.On("UploadProductImage", req.Context(), user.ID, req).Return("https://s3/test.jpg", nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "s3_upload_product_image", "Image uploaded to S3 and URL generated", mock.Anything, mock.Anything).Return()

	cfg.HandlerS3UploadProductImage(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://s3/test.jpg")
	assert.Contains(t, w.Body.String(), "Image URL created successfully (S3)")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerS3UploadProductImage_Error tests the handler's behavior when the S3 upload service returns an error.
// It ensures the handler returns HTTP 500 and logs the error correctly.
func TestHandlerS3UploadProductImage_Error(t *testing.T) {
	mockLogger := new(mockS3Logger)
	mockService := new(mockS3UploadService)
	cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()

	err := errors.New("upload failed")
	mockService.On("UploadProductImage", req.Context(), user.ID, req).Return("", err)
	mockLogger.On("LogHandlerError", mock.Anything, "s3_upload_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()

	cfg.HandlerS3UploadProductImage(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerS3UpdateProductImageByID_Success tests the successful update of a product image by ID in S3 via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns an updated S3 image URL without error.
func TestHandlerS3UpdateProductImageByID_Success(t *testing.T) {
	mockLogger := new(mockS3Logger)
	mockService := new(mockS3UploadService)
	cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	// Patch chiURLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(_ *http.Request, _ string) string {
		return "prod123"
	}
	defer func() { chiURLParam = oldURLParam }()

	req := httptest.NewRequest("POST", "/update/123", nil)
	w := httptest.NewRecorder()

	mockService.On("UpdateProductImage", req.Context(), "prod123", user.ID, req).Return("https://s3/updated.jpg", nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "s3_update_product_image", "Product image updated in S3", mock.Anything, mock.Anything).Return()

	cfg.HandlerS3UpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://s3/updated.jpg")
	assert.Contains(t, w.Body.String(), "Product image updated successfully (S3)")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerS3UpdateProductImageByID_MissingProductID tests the handler's response when the product ID is missing from the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerS3UpdateProductImageByID_MissingProductID(t *testing.T) {
	mockLogger := new(mockS3Logger)
	mockService := new(mockS3UploadService)
	cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	// Patch chiURLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(_ *http.Request, _ string) string {
		return ""
	}
	defer func() { chiURLParam = oldURLParam }()

	req := httptest.NewRequest("POST", "/update", nil)
	w := httptest.NewRecorder()

	mockLogger.On("LogHandlerError", mock.Anything, "s3_update_product_image", "missing_product_id", "Product ID not found", mock.Anything, mock.Anything, nil).Return()

	cfg.HandlerS3UpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Product ID not found")
	mockLogger.AssertExpectations(t)
}

// TestHandlerS3UpdateProductImageByID_Error tests the handler's behavior when the S3 update service returns an error during image update by ID.
// It ensures the handler returns HTTP 500 and logs the error correctly.
func TestHandlerS3UpdateProductImageByID_Error(t *testing.T) {
	mockLogger := new(mockS3Logger)
	mockService := new(mockS3UploadService)
	cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	// Patch chiURLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(_ *http.Request, _ string) string {
		return "prod123"
	}
	defer func() { chiURLParam = oldURLParam }()

	req := httptest.NewRequest("POST", "/update/123", nil)
	w := httptest.NewRecorder()

	err := errors.New("update failed")
	mockService.On("UpdateProductImage", req.Context(), "prod123", user.ID, req).Return("", err)
	mockLogger.On("LogHandlerError", mock.Anything, "s3_update_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()

	cfg.HandlerS3UpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
