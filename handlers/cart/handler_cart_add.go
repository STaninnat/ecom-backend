// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_cart_add.go: Provides HTTP handlers to add items to authenticated user carts and guest session carts.

// getSessionIDFromRequest testable indirection for session ID extraction
var getSessionIDFromRequest = utils.GetSessionIDFromRequest

// HandlerAddItemToUserCart handles HTTP requests to add an item to a user's cart.
// @Summary      Add item to user cart
// @Description  Adds an item to the authenticated user's cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        item  body  CartItemRequest  true  "Cart item payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/cart/items [post]
func (cfg *HandlersCartConfig) HandlerAddItemToUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	cfg.handleCartItemOperation(
		w, r, user.ID, "User ID is required",
		func(r *http.Request) (string, string, int, error) {
			var req CartItemRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return "", "", 0, err
			}
			return req.ProductID, "", req.Quantity, nil
		},
		cfg.GetCartService().AddItemToUserCart,
		"add_item_to_cart",
		"Added item to cart",
		"Item added to cart",
	)
}

// HandlerAddItemToGuestCart handles HTTP requests to add an item to a guest cart (session-based).
// @Summary      Add item to guest cart
// @Description  Adds an item to the guest cart (session-based)
// @Tags         guest-cart
// @Accept       json
// @Produce      json
// @Param        item  body  CartItemRequest  true  "Cart item payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/guest-cart/items [post]
func (cfg *HandlersCartConfig) HandlerAddItemToGuestCart(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionIDFromRequest(r)
	cfg.handleCartItemOperation(
		w, r, sessionID, "Missing session ID",
		func(r *http.Request) (string, string, int, error) {
			var req CartItemRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return "", "", 0, err
			}
			return req.ProductID, "", req.Quantity, nil
		},
		cfg.GetCartService().AddItemToGuestCart,
		"add_item_guest_cart",
		"Added item to guest cart",
		"Item added to cart",
	)
}
