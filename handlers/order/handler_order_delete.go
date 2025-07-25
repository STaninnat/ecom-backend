// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_order_delete.go: Handles HTTP DELETE request to delete an order by ID. Validates request, calls service, logs event, and responds.

// HandlerDeleteOrder handles HTTP DELETE requests to delete an order by its ID.
// @Summary      Delete order
// @Description  Deletes an order by its ID (admin only)
// @Tags         orders
// @Produce      json
// @Param        order_id  path  string  true  "Order ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/orders/{order_id} [delete]
func (cfg *HandlersOrderConfig) HandlerDeleteOrder(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Extract order ID from URL parameters
	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		// Log error for missing order ID
		cfg.Logger.LogHandlerError(
			ctx,
			"delete_order",
			"missing_order_id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	// Call business logic service to delete the order
	err := cfg.GetOrderService().DeleteOrder(ctx, orderID)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "delete_order", ip, userAgent)
		return
	}

	// Log successful deletion with user context
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "delete_order", "Deleted order successful", ip, userAgent)

	// Respond with success message
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Order deleted successfully",
	})
}
