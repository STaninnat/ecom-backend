// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_create.go: Provides HTTP handler for creating categories.

// HandlerCreateCategory handles HTTP POST requests to create a new category.
// Parses the request body for category parameters, validates them, and delegates creation to the category service.
// On success, logs the event and responds with a confirmation message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCategoryConfig) HandlerCreateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryRequest(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		cfg.handleCategoryError,
		"create_category",
		func(ctx context.Context, service CategoryService, params CategoryRequest) (string, error) {
			return service.CreateCategory(ctx, params)
		},
		"Category created successfully",
		http.StatusCreated,
	)
}
