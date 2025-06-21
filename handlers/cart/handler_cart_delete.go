package cart

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

type RemoveItemReq struct {
	ProductID string `json:"product_id"`
}

func (apicfg *HandlersCartConfig) HandlerRemoveItemFromUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req RemoveItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"remove_item_to_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := apicfg.CartMG.RemoveItemFromCart(ctx, user.ID, req.ProductID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"remove_item_from_cart",
			"failed to remove item",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove item from cart")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "remove_item_from_cart", "Removed item from cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item removed from cart",
	})
}

func (apicfg *HandlersCartConfig) HandlerClearUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	err := apicfg.CartMG.ClearCart(ctx, user.ID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"clear_cart",
			"failed to clear cart",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "clear_cart", "Cart cleared", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Cart cleared"})
}

func (apicfg *HandlersCartConfig) HandlerRemoveItemFromGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"remove_item_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req RemoveItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"remove_item_guest_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := apicfg.RemoveGuestItem(ctx, sessionID, req.ProductID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"remove_item_guest_cart",
			"remove guest item failed",
			"Error to remove guest item",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove item")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "remove_item_guest_cart", "Removed item from guest cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item removed from cart",
	})
}

func (apicfg *HandlersCartConfig) HandlerClearGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"clear_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	err := apicfg.DeleteGuestCart(ctx, sessionID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"clear_guest_cart",
			"failed to clear cart",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to clear guest cart")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "clear_guest_cart", "Guest cart cleared", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Guest cart cleared",
	})
}
