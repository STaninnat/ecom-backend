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

	var params CategoryWithIDRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"delete category",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" {
		apicfg.LogHandlerError(
			r.Context(),
			"delete category",
			"missing category id",
			"ID of category is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "ID and Name are required")
		return
	}

	err := apicfg.DB.DeleteCategory(r.Context(), params.ID)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"delete category",
			"delete category failed",
			"Error deleting category",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete category")
		return
	}

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "delete category", "Deleted category successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusNoContent, map[string]string{
		"message": "Deleted category successful",
	})
}
