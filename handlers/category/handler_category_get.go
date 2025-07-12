package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (cfg *HandlersCategoryConfig) HandlerGetAllCategories(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get the category service
	categoryService := cfg.GetCategoryService()

	// Call the service to get all categories
	categories, err := categoryService.GetAllCategories(ctx)
	if err != nil {
		cfg.handleCategoryError(w, r, err, "get_all_categories", ip, userAgent)
		return
	}

	// Log success
	userID := ""
	if user != nil {
		userID = user.ID
	}
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.LogHandlerSuccess(ctxWithUserID, "get_all_categories", "Categories fetched successfully", ip, userAgent)

	// Return categories
	middlewares.RespondWithJSON(w, http.StatusOK, categories)
}
