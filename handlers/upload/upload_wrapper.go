// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"errors"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
)

// upload_wrapper.go: Provides configuration, error handling, and response structures for local and S3 upload handlers.

// HandlersUploadConfig holds dependencies and configuration for local upload handlers.
// Includes the logger, upload path, and the upload service for local file storage operations.
type HandlersUploadConfig struct {
	Config     *handlers.Config
	Logger     handlers.HandlerLogger
	UploadPath string
	Service    UploadService
}

// HandlersUploadS3Config holds dependencies and configuration for S3 upload handlers.
// Includes the logger, upload path, and the upload service for S3 cloud storage operations.
type HandlersUploadS3Config struct {
	Config     *handlers.Config
	Logger     handlers.HandlerLogger
	UploadPath string
	Service    UploadService
}

// imageUploadResponse is the response payload for image upload endpoints.
// Contains success message and the generated image URL for client consumption.
type imageUploadResponse struct {
	Message  string `json:"message"`
	ImageURL string `json:"image_url"`
}

// chiURLParam is a patchable reference to chi.URLParam for testing.
// Allows dependency injection for URL parameter extraction in test scenarios.
var chiURLParam = chi.URLParam

// handleUploadError centralizes error handling for local upload endpoints.
// Logs the error and sends the appropriate HTTP response based on the error type.
// Maps AppError codes to corresponding HTTP status codes and user-friendly messages.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - err: error to handle
//   - operation: string describing the operation
//   - ip: string client IP address
//   - userAgent: string client user agent
func (cfg *HandlersUploadConfig) handleUploadError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	var appErr *handlers.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "missing_product_id":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "invalid_form", "invalid_image":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "db_error", "file_save_failed", "transaction_error", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// handleUploadError centralizes error handling for S3 upload endpoints.
// Logs the error and sends the appropriate HTTP response based on the error type.
// Maps AppError codes to corresponding HTTP status codes and user-friendly messages.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - err: error to handle
//   - operation: string describing the operation
//   - ip: string client IP address
//   - userAgent: string client user agent
func (cfg *HandlersUploadS3Config) handleUploadError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	appErr := &handlers.AppError{}
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "missing_product_id":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "invalid_form", "invalid_image":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "db_error", "file_save_failed", "transaction_error", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
