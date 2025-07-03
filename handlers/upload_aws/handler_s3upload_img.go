package uploadawshandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
)

func (apicfg *HandlersUploadAWSConfig) HandlersUploadProductImageS3(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(r)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_upload-s3",
			"invalid form",
			err.Error(),
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid form or image not found")
		return
	}
	defer file.Close()

	uploader := &utilsuploaders.S3Uploader{
		Client:     apicfg.S3Client,
		BucketName: apicfg.S3Bucket,
	}

	_, imageURL, err := uploader.UploadFileToS3(ctx, file, fileHeader)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"product_image_upload-s3",
			"upload file failed",
			err.Error(),
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to upload image")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "product_image_upload-s3", "Image uploaded successfully and URL generated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadS3Response{
		Message:  "Image uploaded successfully (S3)",
		ImageURL: imageURL,
	})
}
