// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_update.go: Provides HTTP handler for updating categories.

// HandlerUpdateCategory handles HTTP PUT requests to update a category.
// Parses the request body for category parameters, validates them, and delegates update to the category service.
// On success, logs the event and responds with a confirmation message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCategoryConfig) HandlerUpdateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryRequest(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		cfg.handleCategoryError,
		"update_category",
		func(ctx context.Context, service CategoryService, params CategoryRequest) (string, error) {
			return "", service.UpdateCategory(ctx, params)
		},
		"Category updated successfully",
		http.StatusOK,
	)
}
