package userhandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerPromoteUserToAdmin handles HTTP POST requests to promote a user to admin role (admin only).
// Extracts user from request context, validates the target user ID, delegates to user service,
// and handles specific error cases with appropriate HTTP status codes. On success, logs the event
// and responds with success message; on error, logs and returns appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data with user in context and target user ID in body
func (cfg *HandlersUserConfig) HandlerPromoteUserToAdmin(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	user, ok := ctx.Value(contextKeyUser).(database.User)
	if !ok {
		cfg.Logger.LogHandlerError(ctx, "promote_admin", "user_not_found", "User not found in context", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		cfg.Logger.LogHandlerError(ctxWithUserID, "promote_admin", "invalid_request", "Invalid request payload", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := cfg.GetUserService().PromoteUserToAdmin(ctxWithUserID, user, req.UserID)
	if err != nil {
		if appErr, ok := err.(*handlers.AppError); ok {
			switch appErr.Code {
			case "unauthorized_user":
				cfg.Logger.LogHandlerError(ctxWithUserID, "promote_admin", appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
				middlewares.RespondWithError(w, http.StatusForbidden, appErr.Message)
				return
			case "already_admin":
				cfg.Logger.LogHandlerError(ctxWithUserID, "promote_admin", appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
				middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
				return
			case "user_not_found":
				cfg.Logger.LogHandlerError(ctxWithUserID, "promote_admin", appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
				middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
				return
			}
		}
		cfg.handleUserError(w, r, err, "promote_admin", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "promote_admin", "User promoted to admin success", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "User promoted to admin",
	})
}
