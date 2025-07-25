// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_s3.go: Handles S3 product image upload and update with size limits, service calls, logging, and JSON responses.

// HandlerS3UploadProductImage handles HTTP POST requests to upload a new product image to S3 storage.
func (cfg *HandlersUploadS3Config) HandlerS3UploadProductImage(w http.ResponseWriter, r *http.Request, user database.User) {
	handleProductImageUpload(
		w, r, user,
		cfg.Service.UploadProductImage,
		cfg.handleUploadError,
		cfg.Logger,
		"s3_upload_product_image",
		"Image uploaded to S3 and URL generated",
		"Image URL created successfully (S3)",
	)
}

// HandlerS3UpdateProductImageByID handles HTTP POST requests to update a product image by its ID in S3 storage.
func (cfg *HandlersUploadS3Config) HandlerS3UpdateProductImageByID(w http.ResponseWriter, r *http.Request, user database.User) {
	handleUpdateProductImageByID(
		w, r, user,
		func(ctx context.Context, userID string, r *http.Request) (string, error) {
			productID := chiURLParam(r, "id")
			return cfg.Service.UpdateProductImage(ctx, productID, userID, r)
		},
		cfg.handleUploadError,
		cfg.Logger,
		"s3_update_product_image",
		"Product image updated in S3",
		"Product image updated successfully (S3)",
	)
}
