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
// @Summary      Checkout user cart
// @Description  Checks out the authenticated user's cart and creates an order
// @Tags         cart
// @Produce      json
// @Success      200  {object}  CartResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/checkout [post]
func (cfg *HandlersCartConfig) HandlerCheckoutUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	result, err := cfg.GetCartService().CheckoutUserCart(ctx, user.ID)
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

// GuestCheckoutRequest represents the payload for guest cart checkout.
type GuestCheckoutRequest struct {
	UserID string `json:"user_id"`
}

// HandlerCheckoutGuestCart handles HTTP requests to checkout a guest cart (session-based).
// @Summary      Checkout guest cart
// @Description  Checks out the guest cart (session-based) and creates an order
// @Tags         guest-cart
// @Accept       json
// @Produce      json
// @Param        checkout  body  GuestCheckoutRequest  true  "Guest checkout payload"
// @Success      200  {object}  CartResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/checkout [post]
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
	var req GuestCheckoutRequest
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

	result, err := cfg.GetCartService().CheckoutGuestCart(ctx, sessionID, req.UserID)
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
