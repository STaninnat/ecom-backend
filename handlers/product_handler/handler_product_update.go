package producthandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersProductConfig) HandlerUpdateProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_product",
			"invalid body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" || params.CategoryID == "" || params.Name == "" || params.Price <= 0 || params.Stock < 0 {
		apicfg.LogHandlerError(
			ctx,
			"update_product",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields")
		return
	}

	isActive := true
	if params.IsActive != nil {
		isActive = *params.IsActive
	}

	err := apicfg.DB.UpdateProduct(ctx, database.UpdateProductParams{
		ID:          params.ID,
		CategoryID:  utils.ToNullString(params.CategoryID),
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		Price:       fmt.Sprintf("%.2f", params.Price),
		Stock:       params.Stock,
		ImageUrl:    utils.ToNullString(params.ImageURL),
		IsActive:    isActive,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_product",
			"update failed",
			"Error updating product",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't update product")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "update_product", "Updated product successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Product updated successfully",
	})
}
