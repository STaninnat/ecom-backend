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

// handler_product_delete.go: Handles deleting a product by ID: validates input, calls service, logs result, and sends JSON response.

// HandlerDeleteProduct handles HTTP DELETE requests to remove a product by its ID.
// Extracts the product ID from the URL, validates it, and delegates deletion to the product service.
// On success, logs the event and responds with a success message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersProductConfig) HandlerDeleteProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"delete_product",
			"invalid_request",
			"Product ID is required",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	err := cfg.GetProductService().DeleteProduct(ctx, productID)
	if err != nil {
		cfg.handleProductError(w, r, err, "delete_product", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "delete_product", "Delete success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Product deleted successfully",
	})
}
