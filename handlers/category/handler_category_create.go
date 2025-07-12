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

// HandlerCreateCategory handles HTTP POST requests to create a new category.
// It parses the request body for category parameters, validates them, and delegates creation to the category service.
// On success, it logs the event and responds with a confirmation message; on error, it logs and returns the appropriate error response.
func (cfg *HandlersCategoryConfig) HandlerCreateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.LogHandlerError(
			ctx,
			"create_category",
			"invalid_request_body",
			"Failed to parse request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get the category service
	categoryService := cfg.GetCategoryService()

	// Call the service to create the category
	_, err := categoryService.CreateCategory(ctx, params)
	if err != nil {
		cfg.handleCategoryError(w, r, err, "create_category", ip, userAgent)
		return
	}

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.LogHandlerSuccess(ctxWithUserID, "create_category", "Category created successfully", ip, userAgent)

	// Return success response
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Category created successfully",
	})
}
