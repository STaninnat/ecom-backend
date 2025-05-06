package producthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersProductConfig) HandlerDeleteProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		apicfg.LogHandlerError(
			ctx,
			"delete_product",
			"missing product id",
			"Product ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := apicfg.DB.GetProductByID(ctx, productID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"delete_product",
			"get product by ID failed",
			"Product not found",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	if product.ImageUrl.Valid && product.ImageUrl.String != "" {
		_ = utilsuploaders.DeleteFileIfExists(product.ImageUrl.String)
		_ = utilsuploaders.DeleteFileFromS3IfExists(apicfg.S3Client, apicfg.S3Bucket, product.ImageUrl.String)
	}

	err = apicfg.DB.DeleteProductByID(ctx, productID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"delete_product",
			"deletion failed",
			"Error deleting product",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "delete_product", "Delete success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Product deleted successfully",
	})
}
