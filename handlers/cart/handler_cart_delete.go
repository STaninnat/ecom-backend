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

type DeleteItemRequest struct {
	ProductID string `json:"product_id"`
}

// HandlerRemoveItemFromUserCart handles HTTP requests to remove an item from a user's cart.
// It parses and validates the request body, calls the service layer to remove the item,
// logs the operation, and returns an appropriate JSON response or error.
// Expects a valid user and DeleteItemRequest in the request body.
func (cfg *HandlersCartConfig) HandlerRemoveItemFromUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req DeleteItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"remove_item_from_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"remove_item_from_cart",
			"missing product ID",
			"Product ID is required",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	cartService := cfg.GetCartService()
	if err := cartService.RemoveItem(ctx, user.ID, req.ProductID); err != nil {
		cfg.handleCartError(w, r, err, "remove_item_from_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "remove_item_from_cart", "Removed item from cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item removed from cart",
	})
}

// HandlerClearUserCart handles HTTP requests to clear all items from a user's cart.
// It calls the service layer to clear the cart, logs the operation, and returns
// a JSON response indicating success or an error if the operation fails.
// Expects a valid user in the request context.
func (cfg *HandlersCartConfig) HandlerClearUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	cartService := cfg.GetCartService()
	if err := cartService.DeleteUserCart(ctx, user.ID); err != nil {
		cfg.handleCartError(w, r, err, "clear_cart", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "clear_cart", "Cart cleared", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Cart cleared",
	})
}

// HandlerRemoveItemFromGuestCart handles HTTP requests to remove an item from a guest cart (session-based).
// It extracts the session ID, parses and validates the request body, calls the service layer to remove the item,
// logs the operation, and returns an appropriate JSON response or error.
// Expects a valid session ID and DeleteItemRequest in the request body.
func (cfg *HandlersCartConfig) HandlerRemoveItemFromGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"remove_item_from_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req DeleteItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"remove_item_from_guest_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"remove_item_from_guest_cart",
			"missing product ID",
			"Product ID is required",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	cartService := cfg.GetCartService()
	if err := cartService.RemoveGuestItem(ctx, sessionID, req.ProductID); err != nil {
		cfg.handleCartError(w, r, err, "remove_item_from_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "remove_item_from_guest_cart", "Removed item from guest cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item removed from cart",
	})
}

// HandlerClearGuestCart handles HTTP requests to clear all items from a guest cart (session-based).
// It extracts the session ID, calls the service layer to clear the cart, logs the operation,
// and returns a JSON response indicating success or an error if the operation fails.
// Expects a valid session ID in the request context.
func (cfg *HandlersCartConfig) HandlerClearGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"clear_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	cartService := cfg.GetCartService()
	if err := cartService.DeleteGuestCart(ctx, sessionID); err != nil {
		cfg.handleCartError(w, r, err, "clear_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "clear_guest_cart", "Guest cart cleared", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Guest cart cleared",
	})
}
