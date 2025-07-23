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
// Parses the request body for filter parameters, validates them, and delegates filtering to the product service.
// On success, logs the event and responds with the filtered products; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: *database.User representing the authenticated user (may be nil)
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
