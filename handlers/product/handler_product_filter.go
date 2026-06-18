// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_product_filter.go: Handles filtering products: parses filter params, calls service, logs result, and returns matching products.

// HandlerFilterProducts handles HTTP POST requests to filter products based on provided criteria.
// @Summary      Filter products
// @Description  Filters products based on provided criteria
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        category_id  query  string   false  "Category ID"
// @Param        is_active    query  boolean  false  "Is Active"
// @Param        min_price    query  number   false  "Minimum Price"
// @Param        max_price    query  number   false  "Maximum Price"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/filter [get]
func (cfg *HandlersProductConfig) HandlerFilterProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params FilterProductsRequest
	query := r.URL.Query()

	if catID := query.Get("category_id"); catID != "" {
		params.CategoryID.Valid = true
		params.CategoryID.String = catID
	}
	if isActiveStr := query.Get("is_active"); isActiveStr != "" {
		if isActiveStr == "true" {
			params.IsActive.Valid = true
			params.IsActive.Bool = true
		} else if isActiveStr == "false" {
			params.IsActive.Valid = true
			params.IsActive.Bool = false
		}
	}
	if minPriceStr := query.Get("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			params.MinPrice.Valid = true
			params.MinPrice.Float64 = minPrice
		} else {
			cfg.Logger.LogHandlerError(ctx, "filter_products", "invalid_request", "Invalid min_price", ip, userAgent, err)
			middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid min_price")
			return
		}
	}
	if maxPriceStr := query.Get("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			params.MaxPrice.Valid = true
			params.MaxPrice.Float64 = maxPrice
		} else {
			cfg.Logger.LogHandlerError(ctx, "filter_products", "invalid_request", "Invalid max_price", ip, userAgent, err)
			middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid max_price")
			return
		}
	}

	products, err := cfg.GetProductService().FilterProducts(ctx, params)
	if err != nil {
		cfg.handleProductError(w, r, err, "filter_products", ip, userAgent)
		return
	}

	productResp := struct {
		Products []database.Product `json:"products"`
	}{
		Products: products,
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "filter_products", "Filter products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, productResp)
}
