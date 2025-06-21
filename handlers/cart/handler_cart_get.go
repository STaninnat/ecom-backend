package cart

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersCartConfig) HandlerGetUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	cart, err := apicfg.CartMG.GetCartByUserID(ctx, user.ID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_cart",
			"get cart failed",
			"Error getting cart",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to get cart")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "get_cart", "Got user cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}

func (apicfg *HandlersCartConfig) HandlerGetGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	cart, err := apicfg.GetGuestCart(ctx, sessionID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_guest_cart",
			"redis get failed",
			"Failed to retrieve guest cart from Redis",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve guest cart")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "get_guest_cart", "Got guest cart successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, cart)
}
