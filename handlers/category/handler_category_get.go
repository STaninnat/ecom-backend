// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"log"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_category_get.go: Provides HTTP handler to retrieve all categories.

// HandlerGetAllCategories handles HTTP GET requests to retrieve all categories.
// @Summary      Get all categories
// @Description  Retrieves all product categories
// @Tags         categories
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/categories/ [get]
func (cfg *HandlersCategoryConfig) HandlerGetAllCategories(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Call the service to get all categories
	categories, err := cfg.GetCategoryService().GetAllCategories(ctx)
	if err != nil {
		log.Printf("HandlerGetAllCategories: error from service: %+v", err)
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
