// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_update.go: Provides HTTP handler for updating categories.

// HandlerUpdateCategory handles HTTP PUT requests to update a category.
// @Summary      Update category
// @Description  Updates an existing product category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body  object{}  true  "Category payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/categories/ [put]
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
