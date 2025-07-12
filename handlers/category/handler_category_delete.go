package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

// HandlerDeleteCategory handles HTTP DELETE requests to delete a category.
// It extracts the category ID from the URL parameters, validates it, and delegates deletion to the category service.
// On success, it logs the event and responds with a confirmation message; on error, it logs and returns the appropriate error response.
func (cfg *HandlersCategoryConfig) HandlerDeleteCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		cfg.LogHandlerError(
			ctx,
			"delete_category",
			"missing_category_id",
			"Category ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Get the category service
	categoryService := cfg.GetCategoryService()

	// Call the service to delete the category
	err := categoryService.DeleteCategory(ctx, categoryID)
	if err != nil {
		cfg.handleCategoryError(w, r, err, "delete_category", ip, userAgent)
		return
	}

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.LogHandlerSuccess(ctxWithUserID, "delete_category", "Category deleted successfully", ip, userAgent)

	// Return success response
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Category deleted successfully",
	})
}
