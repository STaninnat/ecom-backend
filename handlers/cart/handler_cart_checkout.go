// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_cart_checkout.go: Handles cart checkout requests for authenticated users and guests.

// HandlerCheckoutUserCart handles HTTP requests to checkout a user's cart.
// Calls the service layer, logs the operation, and returns a JSON response with the order details or an error.
// Expects a valid user in the request context.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCartConfig) HandlerCheckoutUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	cartService := cfg.GetCartService()
	result, err := cartService.CheckoutUserCart(ctx, user.ID)
	if err != nil {
		cfg.handleCartError(w, r, err, "checkout_user_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "checkout_user_cart", "User cart checked out successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, CartResponse{
		Message: result.Message,
		OrderID: result.OrderID,
	})
}

// HandlerCheckoutGuestCart handles HTTP requests to checkout a guest cart (session-based).
// Extracts the session ID and user ID, validates them, calls the service layer, logs the operation, and returns a JSON response with the order details or an error.
// Expects a valid session ID and user ID in the request body.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
func (cfg *HandlersCartConfig) HandlerCheckoutGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"checkout_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	// Get user ID from request body or context
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"checkout_guest_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.UserID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"checkout_guest_cart",
			"missing user ID",
			"User ID is required for guest checkout",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "User ID is required for guest checkout")
		return
	}

	cartService := cfg.GetCartService()
	result, err := cartService.CheckoutGuestCart(ctx, sessionID, req.UserID)
	if err != nil {
		cfg.handleCartError(w, r, err, "checkout_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "checkout_guest_cart", "Guest cart checked out successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, CartResponse{
		Message: result.Message,
		OrderID: result.OrderID,
	})
}
