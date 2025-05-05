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

	var params DeleteProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"delete product",
			"invalid body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" {
		apicfg.LogHandlerError(
			r.Context(),
			"delete product",
			"missing product id",
			"ID of product is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	err := apicfg.DB.DeleteProductByID(r.Context(), params.ID)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"delete product",
			"deletion failed",
			"Error deleting product",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "delete product", "Delete success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}
