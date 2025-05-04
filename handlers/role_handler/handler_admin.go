package rolehandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersRoleConfig) PromoteUserToAdmin(w http.ResponseWriter, r *http.Request, user database.User) {
	type promoteRequest struct {
		UserID string `json:"user_id"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)

	if user.Role != "admin" {
		apicfg.LogHandlerError(
			r.Context(),
			"promote admin",
			"unauthorized user",
			"User is not admin",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusForbidden, "Admin privileges required")
		return
	}

	var req promoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		apicfg.LogHandlerError(
			r.Context(),
			"promote admin",
			"invalid payload",
			"Failed to decode request",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	targetUser, err := apicfg.DB.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"promote admin",
			"user not found",
			"Target user not found",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Target user not found")
		return
	}
	if targetUser.Role == "admin" {
		apicfg.LogHandlerError(
			r.Context(),
			"promote admin",
			"already admin",
			"Target user is already admin",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "User is already admin")
		return
	}

	err = apicfg.DB.UpdateUserRole(r.Context(), database.UpdateUserRoleParams{
		Role: "admin",
		ID:   req.UserID,
	})
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"promote admin",
			"update error",
			"failed to update user role",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Error promoting user to admin")
		return
	}

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "promote admin", "User promoted to admin success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User promoted to admin",
	})
}
