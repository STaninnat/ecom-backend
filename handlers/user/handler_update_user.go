// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// UpdateUserRequest represents the payload for updating a user profile.
type UpdateUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// handler_update_user.go: Handles authenticated user update requests by validating input, updating via service, and responding with status.

// HandlerUpdateUser handles HTTP PUT requests to update user information.
// @Summary      Update current user
// @Description  Updates the current user's information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body  UpdateUserRequest  true  "User update payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /v1/users/ [put]
func (cfg *HandlersUserConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	user, ok := ctx.Value(contextKeyUser).(database.User)
	if !ok {
		cfg.Logger.LogHandlerError(ctx, "update_user", "user_not_found", "User not found in context", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	var params *UpdateUserRequest
	var err error
	params, err = auth.DecodeAndValidate[UpdateUserRequest](w, r)
	if err != nil {
		cfg.Logger.LogHandlerError(ctxWithUserID, "update_user", "invalid_request", "Invalid update payload", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = cfg.GetUserService().UpdateUser(ctxWithUserID, user, UpdateUserParams(*params))
	if err != nil {
		cfg.handleUserError(w, r, err, "update_user", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "update_user", "User info updated", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Updated user info successful",
	})
}
