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

func (apicfg *HandlersProductConfig) HandlerDeleteCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CategoryWithIDRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"delete_category",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" {
		apicfg.LogHandlerError(
			ctx,
			"delete_category",
			"missing category id",
			"ID of category is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "ID and Name are required")
		return
	}

	err := apicfg.DB.DeleteCategory(ctx, params.ID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"delete_category",
			"delete category failed",
			"Error deleting category",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete category")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "delete_category", "Deleted category successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusNoContent, map[string]string{
		"message": "Deleted category successful",
	})
}
