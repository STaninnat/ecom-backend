package uploadhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersUploadConfig) HandlerUploadProductImage(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	file, fileHeader, err := utils.ParseAndGetImageFile(r)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"upload_image_product-local",
			"invalid form",
			err.Error(),
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid form or image not found")
		return
	}
	defer file.Close()

	filename, err := utils.SaveUploadedFile(file, fileHeader)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"upload_image_product-local",
			"file save failed",
			err.Error(),
			ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	imageURL := "/static/" + filename

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "upload_image_product-local", "Image uploaded successfully and URL generated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, imageUploadResponse{
		Message:  "Image URL created successfully",
		ImageURL: imageURL,
	})
}
