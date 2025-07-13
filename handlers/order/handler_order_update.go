package orderhandlers

import (
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
)

// HandlerUpdateOrderStatus handles HTTP PUT/PATCH requests to update an order's status.
// It validates the request payload, extracts the order ID from URL parameters,
// calls the business logic service to update the order status, logs the event,
// and responds with a success message. On error or invalid payload,
// it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersOrderConfig) HandlerUpdateOrderStatus(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Parse and validate request payload
	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log error for invalid request payload
		cfg.Logger.LogHandlerError(
			ctx,
			"update_order_status",
			"invalid_request",
			"Failed to parse request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Extract order ID from URL parameters
	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		// Log error for missing order ID
		cfg.Logger.LogHandlerError(
			ctx,
			"update_order_status",
			"missing_order_id",
			"Order ID must be provided",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	// Call business logic service to update order status
	err := cfg.GetOrderService().UpdateOrderStatus(ctx, orderID, req.Status)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "update_order_status", ip, userAgent)
		return
	}

	// Log successful status update
	cfg.Logger.LogHandlerSuccess(ctx, "update_order_status", "Order status updated successfully", ip, userAgent)

	// Respond with success message
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Order status updated successfully",
	})
}
