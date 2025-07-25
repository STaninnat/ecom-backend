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

// handler_cart_delete.go: Provides handlers for managing items in authenticated user and guest carts.

// DeleteItemRequest represents a request containing a product ID.
type DeleteItemRequest struct {
	ProductID string `json:"product_id"`
}

// HandlerRemoveItemFromUserCart handles HTTP requests to remove an item from a user's cart.
// @Summary      Remove item from user cart
// @Description  Removes an item from the authenticated user's cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        item  body  DeleteItemRequest  true  "Delete item payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/items [delete]
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

	if err := cfg.GetCartService().RemoveItem(ctx, user.ID, req.ProductID); err != nil {
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
// @Summary      Clear user cart
// @Description  Clears all items from the authenticated user's cart
// @Tags         cart
// @Produce      json
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/ [delete]
func (cfg *HandlersCartConfig) HandlerClearUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	if err := cfg.GetCartService().DeleteUserCart(ctx, user.ID); err != nil {
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
// @Summary      Remove item from guest cart
// @Description  Removes an item from the guest cart (session-based)
// @Tags         guest-cart
// @Accept       json
// @Produce      json
// @Param        item  body  DeleteItemRequest  true  "Delete item payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/items [delete]
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

	if err := cfg.GetCartService().RemoveGuestItem(ctx, sessionID, req.ProductID); err != nil {
		cfg.handleCartError(w, r, err, "remove_item_from_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "remove_item_from_guest_cart", "Removed item from guest cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item removed from cart",
	})
}

// HandlerClearGuestCart handles HTTP requests to clear all items from a guest cart (session-based).
// @Summary      Clear guest cart
// @Description  Clears all items from the guest cart (session-based)
// @Tags         guest-cart
// @Produce      json
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/ [delete]
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

	if err := cfg.GetCartService().DeleteGuestCart(ctx, sessionID); err != nil {
		cfg.handleCartError(w, r, err, "clear_guest_cart", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "clear_guest_cart", "Guest cart cleared", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Guest cart cleared",
	})
}
