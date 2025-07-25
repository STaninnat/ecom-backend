// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_product_get.go: Handles retrieving all products or by ID with admin check, logging, and JSON response.

// HandlerGetAllProducts handles HTTP GET requests to retrieve all products.
// @Summary      Get all products
// @Description  Retrieves all products
// @Tags         products
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/ [get]
func (cfg *HandlersProductConfig) HandlerGetAllProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	isAdmin := user != nil && user.Role == "admin"
	products, err := cfg.GetProductService().GetAllProducts(ctx, isAdmin)
	if err != nil {
		cfg.handleProductError(w, r, err, "get_products", ip, userAgent)
		return
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_products", "Get all products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, products)
}

// HandlerGetProductByID handles HTTP GET requests to retrieve a product by its ID.
// @Summary      Get product by ID
// @Description  Retrieves a product by its ID
// @Tags         products
// @Produce      json
// @Param        id  path  string  true  "Product ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/{id} [get]
func (cfg *HandlersProductConfig) HandlerGetProductByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"get_product_by_id",
			"invalid_request",
			"Missing product ID",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	isAdmin := user.Role == "admin"
	product, err := cfg.GetProductService().GetProductByID(ctx, productID, isAdmin)
	if err != nil {
		cfg.handleProductError(w, r, err, "get_product_by_id", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_product_by_id", "Get products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, product)
}
