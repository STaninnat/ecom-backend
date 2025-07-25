// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_local_test.go: Tests upload and update product image handlers for success, error, and missing ID cases, verifying correct HTTP responses and proper logging behavior.

const (
	testProductID = "prod123"
)

func TestHandlerUploadProductImage_Scenarios(t *testing.T) {
	tests := []struct {
		name         string
		cfgFactory   func() (any, *httptest.ResponseRecorder, *http.Request, database.User, *mock.Mock, *mock.Mock)
		handler      func(cfg any, w http.ResponseWriter, r *http.Request, user database.User)
		mockSetup    func(service, logger *mock.Mock, req *http.Request, user database.User)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Local_Success",
			cfgFactory: func() (any, *httptest.ResponseRecorder, *http.Request, database.User, *mock.Mock, *mock.Mock) {
				mockLogger := new(mockLogger)
				mockService := new(mockUploadService)
				cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
				user := database.User{ID: "user123"}
				req := httptest.NewRequest("POST", "/upload", nil)
				w := httptest.NewRecorder()
				return cfg, w, req, user, &mockService.Mock, &mockLogger.Mock
			},
			handler: func(cfg any, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.(*HandlersUploadConfig).HandlerUploadProductImage(w, r, user)
			},
			mockSetup: func(service, logger *mock.Mock, req *http.Request, user database.User) {
				service.On("UploadProductImage", req.Context(), user.ID, req).Return("/static/test.jpg", nil)
				logger.On("LogHandlerSuccess", mock.Anything, "upload_product_image", "Image uploaded successfully and URL generated", mock.Anything, mock.Anything).Return()
			},
			expectedCode: http.StatusOK,
			expectedBody: "/static/test.jpg",
		},
		{
			name: "Local_Error",
			cfgFactory: func() (any, *httptest.ResponseRecorder, *http.Request, database.User, *mock.Mock, *mock.Mock) {
				mockLogger := new(mockLogger)
				mockService := new(mockUploadService)
				cfg := &HandlersUploadConfig{Logger: mockLogger, Service: mockService}
				user := database.User{ID: "user123"}
				req := httptest.NewRequest("POST", "/upload", nil)
				w := httptest.NewRecorder()
				return cfg, w, req, user, &mockService.Mock, &mockLogger.Mock
			},
			handler: func(cfg any, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.(*HandlersUploadConfig).HandlerUploadProductImage(w, r, user)
			},
			mockSetup: func(service, logger *mock.Mock, req *http.Request, user database.User) {
				err := errors.New("upload failed")
				service.On("UploadProductImage", req.Context(), user.ID, req).Return("", err)
				logger.On("LogHandlerError", mock.Anything, "upload_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Internal server error",
		},
		{
			name: "S3_Success",
			cfgFactory: func() (any, *httptest.ResponseRecorder, *http.Request, database.User, *mock.Mock, *mock.Mock) {
				mockLogger := new(mockS3Logger)
				mockService := new(mockS3UploadService)
				cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
				user := database.User{ID: "user123"}
				req := httptest.NewRequest("POST", "/upload", nil)
				w := httptest.NewRecorder()
				return cfg, w, req, user, &mockService.Mock, &mockLogger.Mock
			},
			handler: func(cfg any, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.(*HandlersUploadS3Config).HandlerS3UploadProductImage(w, r, user)
			},
			mockSetup: func(service, logger *mock.Mock, req *http.Request, user database.User) {
				service.On("UploadProductImage", req.Context(), user.ID, req).Return("https://s3/test.jpg", nil)
				logger.On("LogHandlerSuccess", mock.Anything, "s3_upload_product_image", "Image uploaded to S3 and URL generated", mock.Anything, mock.Anything).Return()
			},
			expectedCode: http.StatusOK,
			expectedBody: "https://s3/test.jpg",
		},
		{
			name: "S3_Error",
			cfgFactory: func() (any, *httptest.ResponseRecorder, *http.Request, database.User, *mock.Mock, *mock.Mock) {
				mockLogger := new(mockS3Logger)
				mockService := new(mockS3UploadService)
				cfg := &HandlersUploadS3Config{Logger: mockLogger, Service: mockService}
				user := database.User{ID: "user123"}
				req := httptest.NewRequest("POST", "/upload", nil)
				w := httptest.NewRecorder()
				return cfg, w, req, user, &mockService.Mock, &mockLogger.Mock
			},
			handler: func(cfg any, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.(*HandlersUploadS3Config).HandlerS3UploadProductImage(w, r, user)
			},
			mockSetup: func(service, logger *mock.Mock, req *http.Request, user database.User) {
				err := errors.New("upload failed")
				service.On("UploadProductImage", req.Context(), user.ID, req).Return("", err)
				logger.On("LogHandlerError", mock.Anything, "s3_upload_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, w, req, user, service, logger := tt.cfgFactory()
			tt.mockSetup(service, logger, req, user)
			tt.handler(cfg, w, req, user)
			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			service.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
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

	req = req.WithContext(context.WithValue(req.Context(), contextKey("chi.URLParams"), map[string]string{"id": testProductID}))
	mockService.On("UpdateProductImage", req.Context(), testProductID, user.ID, req).Return("/static/updated.jpg", nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_product_image", "Product image updated", mock.Anything, mock.Anything).Return()

	// Patch chi.URLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(_ *http.Request, _ string) string {
		return testProductID
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
	chiURLParam = func(_ *http.Request, _ string) string {
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

	req = req.WithContext(context.WithValue(req.Context(), contextKey("chi.URLParams"), map[string]string{"id": testProductID}))
	err := errors.New("update failed")
	mockService.On("UpdateProductImage", req.Context(), testProductID, user.ID, req).Return("", err)
	mockLogger.On("LogHandlerError", mock.Anything, "update_product_image", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, err).Return()

	// Patch chi.URLParam for test
	oldURLParam := chiURLParam
	chiURLParam = func(_ *http.Request, _ string) string {
		return testProductID
	}
	defer func() { chiURLParam = oldURLParam }()

	cfg.HandlerUpdateProductImageByID(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
