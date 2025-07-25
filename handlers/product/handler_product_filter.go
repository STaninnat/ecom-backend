// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"context"
	"encoding/json"
	"net/http"

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
// @Param        filter  body  object{}  true  "Filter payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/filter [get]
func (cfg *HandlersProductConfig) HandlerFilterProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params FilterProductsRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"filter_products",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
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
