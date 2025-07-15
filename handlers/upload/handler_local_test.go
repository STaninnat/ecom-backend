package uploadhandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerUploadProductImage_Success tests the successful upload of a product image via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns an image URL without error.
func TestHandlerUploadProductImage_Success(t *testing.T) {
	mockLogger := new(mockLogger)
	mockService := new(mockUploadService)
	cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()

	mockService.On("UploadProductImage", req.Context(), user.ID, req).Return("/static/test.jpg", nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "upload_product_image", "Image uploaded successfully and URL generated", mock.Anything, mock.Anything).Return()

	cfg.HandlerUploadProductImage(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/static/test.jpg")
	assert.Contains(t, w.Body.String(), "Image URL created successfully")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUploadProductImage_Error tests the handler's behavior when the upload service returns an error.
// It ensures the handler returns HTTP 500 and logs the error correctly.
func TestHandlerUploadProductImage_Error(t *testing.T) {
	mockLogger := new(mockLogger)
	mockService := new(mockUploadService)
	cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()

	err := errors.New("upload failed")
	mockService.On("UploadProductImage", req.Context(), user.ID, req).Return("", err)
	mockLogger.On("LogHandlerError", mock.Anything, "upload_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()

	cfg.HandlerUploadProductImage(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateProductImageByID_Success tests the successful update of a product image by ID via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns an updated image URL without error.
func TestHandlerUpdateProductImageByID_Success(t *testing.T) {
	mockLogger := new(mockLogger)
	mockService := new(mockUploadService)
	cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/update/123", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(context.WithValue(req.Context(), contextKey("chi.URLParams"), map[string]string{"id": "prod123"}))
	mockService.On("UpdateProductImage", req.Context(), "prod123", user.ID, req).Return("/static/updated.jpg", nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_product_image", "Product image updated", mock.Anything, mock.Anything).Return()

	// Patch chi.URLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(r *http.Request, key string) string {
		return "prod123"
	}
	defer func() { chiURLParam = oldURLParam }()

	cfg.HandlerUpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/static/updated.jpg")
	assert.Contains(t, w.Body.String(), "Product image updated successfully")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateProductImageByID_MissingProductID tests the handler's response when the product ID is missing from the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerUpdateProductImageByID_MissingProductID(t *testing.T) {
	mockLogger := new(mockLogger)
	mockService := new(mockUploadService)
	cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/update", nil)
	w := httptest.NewRecorder()

	// Patch chi.URLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(r *http.Request, key string) string {
		return ""
	}
	defer func() { chiURLParam = oldURLParam }()

	mockLogger.On("LogHandlerError", mock.Anything, "update_product_image", "missing_product_id", "Product ID not found", mock.Anything, mock.Anything, nil).Return()

	cfg.HandlerUpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Product ID not found")
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateProductImageByID_Error tests the handler's behavior when the update service returns an error during image update by ID.
// It ensures the handler returns HTTP 500 and logs the error correctly.
func TestHandlerUpdateProductImageByID_Error(t *testing.T) {
	mockLogger := new(mockLogger)
	mockService := new(mockUploadService)
	cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
	user := database.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/update/123", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(context.WithValue(req.Context(), contextKey("chi.URLParams"), map[string]string{"id": "prod123"}))
	err := errors.New("update failed")
	mockService.On("UpdateProductImage", req.Context(), "prod123", user.ID, req).Return("", err)
	mockLogger.On("LogHandlerError", mock.Anything, "update_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()

	// Patch chi.URLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(r *http.Request, key string) string {
		return "prod123"
	}
	defer func() { chiURLParam = oldURLParam }()

	cfg.HandlerUpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
