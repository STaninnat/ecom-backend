// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

// handler_order_get.go: HTTP handlers for fetching orders and order items, with service calls and structured logging.

// HandlerGetAllOrders handles HTTP GET requests to retrieve all orders (admin only).
// Calls the business logic service, logs the event, and responds with the complete order list or error.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user (admin only)
func (cfg *HandlersOrderConfig) HandlerGetAllOrders(w http.ResponseWriter, r *http.Request, _ database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Call business logic service to retrieve all orders
	orders, err := cfg.GetOrderService().GetAllOrders(ctx)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "list_all_orders", ip, userAgent)
		return
	}

	// Log successful retrieval of all orders
	cfg.Logger.LogHandlerSuccess(ctx, "list_all_orders", "Listed all orders", ip, userAgent)

	// Respond with complete order list
	middlewares.RespondWithJSON(w, http.StatusOK, orders)
}

// HandlerGetUserOrders handles HTTP GET requests to retrieve orders for a specific user.
// Calls the business logic service, logs the event, and responds with the user's order list or error.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersOrderConfig) HandlerGetUserOrders(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Call business logic service to retrieve user orders
	orders, err := cfg.GetOrderService().GetUserOrders(ctx, user)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "get_user_orders", ip, userAgent)
		return
	}

	// Log successful retrieval of user orders
	cfg.Logger.LogHandlerSuccess(ctx, "get_user_orders", "Retrieved user orders", ip, userAgent)

	// Respond with user's order list
	middlewares.RespondWithJSON(w, http.StatusOK, orders)
}

// HandlerGetOrderByID handles HTTP GET requests to retrieve a specific order by its ID.
// Extracts the order ID from URL parameters, validates the request, calls the business logic service, logs the event, and responds with the order information or error.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersOrderConfig) HandlerGetOrderByID(w http.ResponseWriter, r *http.Request, user database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Extract order ID from URL parameters
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		// Log error for missing order ID
		cfg.Logger.LogHandlerError(
			ctx,
			"get_order_by_id",
			"missing_order_id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order ID")
		return
	}

	// Call business logic service to retrieve specific order
	order, err := cfg.GetOrderService().GetOrderByID(ctx, orderID, user)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "get_order_by_id", ip, userAgent)
		return
	}

	// Log successful retrieval with user context
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_order_by_id", "Fetched order details", ip, userAgent)

	// Respond with order details
	middlewares.RespondWithJSON(w, http.StatusOK, order)
}

// HandlerGetOrderItemsByOrderID handles HTTP GET requests to retrieve items for a specific order.
// Extracts the order ID from URL parameters, validates the request, calls the business logic service, logs the event, and responds with the order items list or error.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersOrderConfig) HandlerGetOrderItemsByOrderID(w http.ResponseWriter, r *http.Request, _ database.User) {
	// Extract request metadata for logging
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Extract order ID from URL parameters
	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		// Log error for missing order ID
		cfg.Logger.LogHandlerError(
			ctx,
			"get_order_items",
			"missing_order_id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	// Call business logic service to retrieve order items
	items, err := cfg.GetOrderService().GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		// Handle and log any errors from the service layer
		cfg.handleOrderError(w, r, err, "get_order_items", ip, userAgent)
		return
	}

	// Log successful retrieval of order items
	cfg.Logger.LogHandlerSuccess(ctx, "get_order_items", "Retrieved order items", ip, userAgent)

	// Respond with order items list
	middlewares.RespondWithJSON(w, http.StatusOK, items)
}
