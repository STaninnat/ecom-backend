package userhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

type contextKey string

const contextKeyUser contextKey = "user"

// HandlerGetUser returns the user info as JSON
func (cfg *HandlersUserConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	user, ok := ctx.Value(contextKeyUser).(database.User)
	if !ok {
		cfg.Logger.LogHandlerError(ctx, "get_user", "user_not_found", "User not found in context", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	resp, err := cfg.GetUserService().GetUser(ctxWithUserID, user)
	if err != nil {
		cfg.handleUserError(w, r, err, "get_user", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_user", "Get user info success", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, resp)
}
