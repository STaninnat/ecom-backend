package producthandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersProductConfig) HandlerDeleteProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	type DeleteProductRequest struct {
		ID string `json:"id"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params DeleteProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"delete_product",
			"invalid body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" {
		apicfg.LogHandlerError(
			ctx,
			"delete_product",
			"missing product id",
			"ID of product is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	err := apicfg.DB.DeleteProductByID(ctx, params.ID)
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

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}
