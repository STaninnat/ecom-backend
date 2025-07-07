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

// HandlerGetUser returns the user info as JSON
func (cfg *HandlersUserConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)

	resp, err := cfg.GetUserService().GetUser(ctxWithUserID, user)
	if err != nil {
		cfg.handleUserError(w, r, err, "get_user", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_user", "Get user info success", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, resp)
}

// HandlerUpdateUser updates the user's information
func (cfg *HandlersUserConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		cfg.Logger.LogHandlerError(ctxWithUserID, "update_user", "invalid_request", "Failed to parse body", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Name == "" || params.Email == "" {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Name and Email are required")
		return
	}

	err := cfg.GetUserService().UpdateUser(ctxWithUserID, user, UpdateUserParams(params))
	if err != nil {
		cfg.handleUserError(w, r, err, "update_user", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "update_user", "User info updated", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusNoContent, handlers.HandlerResponse{
		Message: "Updated user info successful",
	})
}
