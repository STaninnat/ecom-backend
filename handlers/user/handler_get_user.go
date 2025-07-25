// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_get_user.go: Handles GET user requests by extracting user from context, fetching user data, logging, and responding.

// contextKey defines a type for context keys to ensure type safety.
type contextKey string

// contextKeyUser is the context key used to store user information in request context.
const contextKeyUser contextKey = "user"

// HandlerGetUser handles HTTP GET requests to retrieve user information.
// @Summary      Get current user
// @Description  Retrieves the current user's information
// @Tags         users
// @Produce      json
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  map[string]string
// @Router       /v1/users/ [get]
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
