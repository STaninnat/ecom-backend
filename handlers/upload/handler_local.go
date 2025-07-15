package uploadhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerUploadProductImage handles HTTP POST requests to upload a new product image (local storage).
// It enforces a max upload size, delegates to the upload service, logs the event, and responds with the image URL.
// On error, it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersUploadConfig) HandlerUploadProductImage(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	const maxUploadSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	imageURL, err := cfg.Service.UploadProductImage(ctx, user.ID, r)
	if err != nil {
		cfg.handleUploadError(w, r, err, "upload_product_image", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "upload_product_image", "Image uploaded successfully and URL generated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  "Image URL created successfully",
		ImageURL: imageURL,
	})
}

// HandlerUpdateProductImageByID handles HTTP POST requests to update a product image by its ID (local storage).
// It extracts the product ID from the URL, delegates to the upload service, logs the event, and responds with the updated image URL.
// On error or missing ID, it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersUploadConfig) HandlerUpdateProductImageByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	productID := chiURLParam(r, "id")
	if productID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_product_image",
			"missing_product_id",
			"Product ID not found",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID not found")
		return
	}

	imageURL, err := cfg.Service.UpdateProductImage(ctx, productID, user.ID, r)
	if err != nil {
		cfg.handleUploadError(w, r, err, "update_product_image", ip, userAgent)
		return
	}

	ctxWithUser := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUser, "update_product_image", "Product image updated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  "Product image updated successfully",
		ImageURL: imageURL,
	})
}
