// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_s3.go: Handles S3 product image upload and update with size limits, service calls, logging, and JSON responses.

// HandlerS3UploadProductImage handles HTTP POST requests to upload a new product image to S3 storage.
// Enforces a max upload size, delegates to the S3 upload service, logs the event, and responds with the S3 image URL.
// On error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersUploadS3Config) HandlerS3UploadProductImage(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	const maxUploadSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	imageURL, err := cfg.Service.UploadProductImage(ctx, user.ID, r)
	if err != nil {
		cfg.handleUploadError(w, r, err, "s3_upload_product_image", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "s3_upload_product_image", "Image uploaded to S3 and URL generated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  "Image URL created successfully (S3)",
		ImageURL: imageURL,
	})
}

// HandlerS3UpdateProductImageByID handles HTTP POST requests to update a product image by its ID in S3 storage.
// Extracts the product ID from the URL, delegates to the S3 upload service, logs the event, and responds with the updated S3 image URL.
// On error or missing ID, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersUploadS3Config) HandlerS3UpdateProductImageByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	productID := chiURLParam(r, "id")
	if productID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"s3_update_product_image",
			"missing_product_id",
			"Product ID not found",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID not found")
		return
	}

	imageURL, err := cfg.Service.UpdateProductImage(ctx, productID, user.ID, r)
	if err != nil {
		cfg.handleUploadError(w, r, err, "s3_update_product_image", ip, userAgent)
		return
	}

	ctxWithUser := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUser, "s3_update_product_image", "Product image updated in S3", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  "Product image updated successfully (S3)",
		ImageURL: imageURL,
	})
}
