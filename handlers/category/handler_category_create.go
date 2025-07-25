// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_category_create.go: Provides HTTP handler for creating categories.

// HandlerCreateCategory handles HTTP POST requests to create a new category.
// @Summary      Create category
// @Description  Creates a new product category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body  object{}  true  "Category payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/categories/ [post]
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
