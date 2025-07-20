package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerGetAllCategories handles HTTP GET requests to retrieve all categories.
// Delegates the retrieval to the category service and returns the categories as JSON.
// On success, logs the event and responds with the category list; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: *database.User representing the authenticated user (may be nil)
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
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_all_categories", "Categories fetched successfully", ip, userAgent)

	// Return categories
	middlewares.RespondWithJSON(w, http.StatusOK, categories)
}
