package categoryhandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerUpdateCategory handles HTTP PUT requests to update a category.
// It parses the request body for category parameters, validates them, and delegates update to the category service.
// On success, it logs the event and responds with a confirmation message; on error, it logs and returns the appropriate error response.
func (cfg *HandlersCategoryConfig) HandlerUpdateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.LogHandlerError(
			ctx,
			"update_category",
			"invalid_request_body",
			"Failed to parse request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get the category service
	categoryService := cfg.GetCategoryService()

	// Call the service to update the category
	err := categoryService.UpdateCategory(ctx, params)
	if err != nil {
		cfg.handleCategoryError(w, r, err, "update_category", ip, userAgent)
		return
	}

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.LogHandlerSuccess(ctxWithUserID, "update_category", "Category updated successfully", ip, userAgent)

	// Return success response
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Category updated successfully",
	})
}
