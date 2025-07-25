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

// handler_cart_update.go: Provides handlers for updating cart item quantities for users and guests.

// HandlerUpdateItemQuantity handles HTTP requests to update the quantity of an item in a user's cart.
// @Summary      Update item quantity in user cart
// @Description  Updates the quantity of an item in the authenticated user's cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        item  body  CartUpdateRequest  true  "Cart update payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/items [put]
func (cfg *HandlersCartConfig) HandlerUpdateItemQuantity(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req CartUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_item_quantity",
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
			"update_item_quantity",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	if err := cfg.GetCartService().UpdateItemQuantity(ctx, user.ID, req.ProductID, req.Quantity); err != nil {
		cfg.handleCartError(w, r, err, "update_item_quantity", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "update_item_quantity", "Updated item quantity", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item quantity updated",
	})
}

// HandlerUpdateGuestItemQuantity handles HTTP requests to update the quantity of an item in a guest cart (session-based).
// @Summary      Update item quantity in guest cart
// @Description  Updates the quantity of an item in the guest cart (session-based)
// @Tags         guest-cart
// @Accept       json
// @Produce      json
// @Param        item  body  CartUpdateRequest  true  "Cart update payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/items [put]
func (cfg *HandlersCartConfig) HandlerUpdateGuestItemQuantity(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := getSessionIDFromRequest(r)
	if sessionID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_guest_item_quantity",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req CartUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_guest_item_quantity",
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
			"update_guest_item_quantity",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	if err := cfg.GetCartService().UpdateGuestItemQuantity(ctx, sessionID, req.ProductID, req.Quantity); err != nil {
		cfg.handleCartError(w, r, err, "update_guest_item_quantity", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "update_guest_item_quantity", "Updated guest item quantity", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item quantity updated",
	})
}
