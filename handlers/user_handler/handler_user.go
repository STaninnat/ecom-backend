package userhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersUserConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	userResp := models.MapUserToResponse(user)

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userResp.ID)

	apicfg.LogHandlerSuccess(ctxWithUserID, "get_user", "Get user info success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, userResp)
}

func (apicfg *HandlersUserConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_user",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Name == "" || params.Email == "" {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Name and Email are required")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_user",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdateUserInfo(ctx, database.UpdateUserInfoParams{
		ID:        user.ID,
		Name:      params.Name,
		Email:     params.Email,
		Phone:     utils.ToNullString(params.Phone),
		Address:   utils.ToNullString(params.Address),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_user",
			"update failed",
			"DB update error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update user info")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_user",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "update_user", "User info updated", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusNoContent, handlers.HandlerResponse{
		Message: "Updated user info successful",
	})
}
