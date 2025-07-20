package carthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerGetUserCart handles HTTP requests to retrieve a user's cart.
// Calls the service layer, logs the operation, and returns a JSON response with the cart data or an error.
// Expects a valid user in the request context.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCartConfig) HandlerGetUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	cartService := cfg.GetCartService()
	cart, err := cartService.GetUserCart(ctx, user.ID)
	if err != nil {
		cfg.handleCartError(w, r, err, "get_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_cart", "Got user cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}

// HandlerGetGuestCart handles HTTP requests to retrieve a guest cart (session-based).
// Extracts the session ID, calls the service layer, logs the operation, and returns a JSON response with the cart data or an error.
// Expects a valid session ID in the request context.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
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

	cartService := cfg.GetCartService()
	cart, err := cartService.GetGuestCart(ctx, sessionID)
	if err != nil {
		cfg.handleCartError(w, r, err, "get_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "get_guest_cart", "Got guest cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}
