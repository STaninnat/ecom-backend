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

// handler_cart_add.go: Provides HTTP handlers to add items to authenticated user carts and guest session carts.

// getSessionIDFromRequest testable indirection for session ID extraction
var getSessionIDFromRequest = utils.GetSessionIDFromRequest

// HandlerAddItemToUserCart handles HTTP requests to add an item to a user's cart.
// Parses and validates the request body, calls the service layer, logs the operation, and returns a JSON response or error.
// Expects a valid user and CartItemRequest in the request body.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersCartConfig) HandlerAddItemToUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req CartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 {
		cfg.Logger.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	// Use service layer
	cartService := cfg.GetCartService()
	if err := cartService.AddItemToUserCart(ctx, user.ID, req.ProductID, req.Quantity); err != nil {
		cfg.handleCartError(w, r, err, "add_item_to_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "add_item_to_cart", "Added item to cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item added to cart",
	})
}

// HandlerAddItemToGuestCart handles HTTP requests to add an item to a guest cart (session-based).
// Extracts the session ID, parses and validates the request body, calls the service layer, logs the operation, and returns a JSON response or error.
// Expects a valid session ID and CartItemRequest in the request body.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
func (cfg *HandlersCartConfig) HandlerAddItemToGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req CartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 {
		cfg.Logger.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	// Use service layer
	cartService := cfg.GetCartService()
	if err := cartService.AddItemToGuestCart(ctx, sessionID, req.ProductID, req.Quantity); err != nil {
		cfg.handleCartError(w, r, err, "add_item_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "add_item_guest_cart", "Added item to guest cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item added to cart",
	})
}
