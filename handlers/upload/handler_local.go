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

// handler_local.go: Handles product image upload and update (local storage) with size limits, service delegation, logging, and JSON responses.

// HandlerUploadProductImage handles HTTP POST requests to upload a new product image (local storage).
// @Summary      Upload product image
// @Description  Uploads a new product image (admin only). The backend (local filesystem or S3) is selected by server configuration. The API and request/response are the same regardless of backend.
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Product image file"
// @Success      200  {object}  imageUploadResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/upload-image [post]
func (cfg *HandlersUploadConfig) HandlerUploadProductImage(w http.ResponseWriter, r *http.Request, user database.User) {
	handleProductImageUpload(
		w, r, user,
		cfg.Service.UploadProductImage,
		cfg.handleUploadError,
		cfg.Logger,
		"upload_product_image",
		"Image uploaded successfully and URL generated",
		"Image URL created successfully",
	)
}

// handleUpdateProductImageByID is a shared helper for update-by-ID logic for both local and S3 uploads.
func handleUpdateProductImageByID(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
	serviceUpdate func(ctx context.Context, userID string, r *http.Request) (string, error),
	handleUploadError func(http.ResponseWriter, *http.Request, error, string, string, string),
	logger handlers.HandlerLogger,
	operation, logMsg, respMsg string,
) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	productID := chiURLParam(r, "id")
	if productID == "" {
		logger.LogHandlerError(
			ctx,
			operation,
			"missing_product_id",
			"Product ID not found",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID not found")
		return
	}

	// Wrap the serviceUpdate to inject productID
	wrappedServiceUpdate := func(ctx context.Context, userID string, r *http.Request) (string, error) {
		return serviceUpdate(ctx, userID, r)
	}

	handleProductImageUpload(
		w, r, user,
		wrappedServiceUpdate,
		handleUploadError,
		logger,
		operation,
		logMsg,
		respMsg,
	)
}

// HandlerUpdateProductImageByID handles HTTP POST requests to update a product image by ID (local storage).
// @Summary      Update product image by ID
// @Description  Updates a product image by product ID (admin only). The backend (local filesystem or S3) is selected by server configuration. The API and request/response are the same regardless of backend.
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path      string  true  "Product ID"
// @Param        file  formData  file    true  "Product image file"
// @Success      200  {object}  imageUploadResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/{id}/image [post]
func (cfg *HandlersUploadConfig) HandlerUpdateProductImageByID(w http.ResponseWriter, r *http.Request, user database.User) {
	handleUpdateProductImageByID(
		w, r, user,
		func(ctx context.Context, userID string, r *http.Request) (string, error) {
			productID := chiURLParam(r, "id")
			return cfg.Service.UpdateProductImage(ctx, productID, userID, r)
		},
		cfg.handleUploadError,
		cfg.Logger,
		"update_product_image",
		"Product image updated",
		"Product image updated successfully",
	)
}

// handleProductImageUpload is a shared helper for product image upload/update logic.
func handleProductImageUpload(
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
	serviceUpload func(ctx context.Context, userID string, r *http.Request) (string, error),
	handleUploadError func(http.ResponseWriter, *http.Request, error, string, string, string),
	logger handlers.HandlerLogger,
	operation, logMsg, respMsg string,
) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	const maxUploadSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	imageURL, err := serviceUpload(ctx, user.ID, r)
	if err != nil {
		handleUploadError(w, r, err, operation, ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	logger.LogHandlerSuccess(ctxWithUserID, operation, logMsg, ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  respMsg,
		ImageURL: imageURL,
	})
}
