package producthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersProductConfig) HandlerGetAllCategories(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	categories, err := apicfg.DB.GetAllCategories(ctx)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_all_categories",
			"get categories failed",
			"Error fetching all categories",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_all_categories", "Fetched successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, categories)
}
