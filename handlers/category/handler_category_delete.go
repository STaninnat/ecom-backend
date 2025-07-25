// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_delete.go: Provides HTTP handler for deleting categories by ID.

// HandlerDeleteCategory handles HTTP DELETE requests to delete a category.
// @Summary      Delete category
// @Description  Deletes a product category by ID
// @Tags         categories
// @Produce      json
// @Param        id  path  string  true  "Category ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/categories/{id} [delete]
func (cfg *HandlersCategoryConfig) HandlerDeleteCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryDelete(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		SharedHandleCategoryError,
	)
}
