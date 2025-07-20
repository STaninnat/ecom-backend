package orderhandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerCreateOrder handles HTTP POST requests to create a new order.
// Validates the request payload, parses order creation parameters, calls the business logic service, logs the event, and responds with the created order details or error.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersOrderConfig) HandlerCreateOrder(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Parse and validate request payload
	var params CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		// Log error for invalid request payload
		cfg.Logger.LogHandlerError(
			ctx,
			"create_order",
			"invalid_request",
			"Failed to parse request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Call business logic service to create the order
	result, err := cfg.GetOrderService().CreateOrder(ctx, user, params)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "create_order", ip, userAgent)
		return
	}

	// Log successful order creation with user context
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "create_order", "Created order successful", ip, userAgent)

	// Respond with created order details
	middlewares.RespondWithJSON(w, http.StatusCreated, result)
}
