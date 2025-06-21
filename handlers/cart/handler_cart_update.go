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

type UpdateItemReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (apicfg *HandlersCartConfig) HandlerUpdateItemInUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req UpdateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" || req.Quantity <= 0 {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_item_in_cart",
			"invalid request",
			"Invalid payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Quantity <= 0 {
		err := apicfg.CartMG.RemoveItemFromCart(ctx, user.ID, req.ProductID)
		if err != nil {
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove item")
			return
		}

		middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
			Message: "Item removed from cart",
		})
		return
	}

	err := apicfg.CartMG.UpdateItemQuantity(ctx, user.ID, req.ProductID, req.Quantity)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_item_in_cart",
			"failed to update",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update item")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "update_item_in_cart", "Updated item quantity", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item updated",
	})
}

func (apicfg *HandlersCartConfig) HandlerUpdateGuestCartItem(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_item_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req UpdateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" || req.Quantity <= 0 {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_item_guest_cart",
			"invalid request",
			"Invalid payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quantity <= 0 {
		err := apicfg.RemoveGuestItem(ctx, sessionID, req.ProductID)
		if err != nil {
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove item")
			return
		}

		middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
			Message: "Item removed from cart",
		})
		return
	}

	err := apicfg.UpdateGuestItemQuantity(ctx, sessionID, req.ProductID, req.Quantity)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_item_guest_cart",
			"failed to update",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update item")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "update_item_guest_cart", "Updated guest item quantity", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item updated",
	})
}
