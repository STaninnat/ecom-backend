// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_cart_get.go: Provides handlers to fetch shopping cart data for users and guests.

// HandlerGetUserCart handles HTTP requests to retrieve a user's cart.
// @Summary      Get user cart
// @Description  Retrieves the authenticated user's cart
// @Tags         cart
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/items [get]
func (cfg *HandlersCartConfig) HandlerGetUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	cart, err := cfg.GetCartService().GetUserCart(ctx, user.ID)
	if err != nil {
		cfg.handleCartError(w, r, err, "get_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_cart", "Got user cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}

// HandlerGetGuestCart handles HTTP requests to retrieve a guest cart (session-based).
// @Summary      Get guest cart
// @Description  Retrieves the guest cart (session-based)
// @Tags         guest-cart
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/ [get]
func (cfg *HandlersCartConfig) HandlerGetGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"get_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	cart, err := cfg.GetCartService().GetGuestCart(ctx, sessionID)
	if err != nil {
		cfg.handleCartError(w, r, err, "get_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "get_guest_cart", "Got guest cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}
