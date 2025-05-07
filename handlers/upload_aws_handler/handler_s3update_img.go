package uploadawshandlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersUploadAWSConfig) HandlerUpdateProductImageS3ByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	productID := chi.URLParam(r, "id")
	if productID == "" {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"missing product id",
			"Product ID not found",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	product, err := queries.GetProductByID(ctx, productID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"get product by ID failed",
			"Product not found",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(r)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"invalid form",
			err.Error(),
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid image file")
		return
	}
	defer file.Close()

	uploader := &utilsuploaders.S3Uploader{
		Client:     apicfg.S3Client,
		BucketName: apicfg.S3Bucket,
	}

	if product.ImageUrl.Valid && product.ImageUrl.String != "" {
		_ = utilsuploaders.DeleteFileFromS3IfExists(apicfg.S3Client, apicfg.S3Bucket, product.ImageUrl.String)
	}

	imageURL, err := uploader.UploadFileToS3(ctx, file, fileHeader)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"upload file failed",
			err.Error(),
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to upload image")
		return
	}

	err = queries.UpdateProductImageURL(ctx, database.UpdateProductImageURLParams{
		ID:        productID,
		ImageUrl:  sql.NullString{String: imageURL, Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"db update failed",
			"Failed to update product image_url",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update product image")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_update-s3",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ctxWithUser := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUser, "product_image_update-s3", "Product image updated successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadS3Response{
		Message:  "Product image updated (S3)",
		ImageURL: imageURL,
	})
}
