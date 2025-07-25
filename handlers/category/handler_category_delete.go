// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_delete.go: Provides HTTP handler for deleting categories by ID.

// HandlerDeleteCategory handles HTTP DELETE requests to delete a category.
// Extracts the category ID from the URL parameters, validates it, and delegates deletion to the category service.
// On success, logs the event and responds with a confirmation message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCategoryConfig) HandlerDeleteCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryDelete(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		SharedHandleCategoryError,
	)
}
