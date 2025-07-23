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

// handler_product_create.go: Handles creating a new product: parses request, calls service, logs outcome, and sends JSON response.

// HandlerCreateProduct handles HTTP POST requests to create a new product.
// Parses the request body for product parameters, validates them, and delegates creation to the product service.
// On success, logs the event and responds with the new product ID; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersProductConfig) HandlerCreateProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"create_product",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	id, err := cfg.GetProductService().CreateProduct(ctx, params)
	if err != nil {
		cfg.handleProductError(w, r, err, "create_product", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "create_product", "Created product successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, productResponse{
		Message:   "Product created successfully",
		ProductID: id,
	})
}
