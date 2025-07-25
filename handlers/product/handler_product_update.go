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

// handler_product_update.go: Handles updating a product: parses input, calls service, logs result, and returns success or error response.

// HandlerUpdateProduct handles HTTP PUT requests to update an existing product.
// @Summary      Update product
// @Description  Updates an existing product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body  object{}  true  "Product payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/products/ [put]
func (cfg *HandlersProductConfig) HandlerUpdateProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_product",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := cfg.GetProductService().UpdateProduct(ctx, params)
	if err != nil {
		cfg.handleProductError(w, r, err, "update_product", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "update_product", "Updated product successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Product updated successfully",
	})
}
