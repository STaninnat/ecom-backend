package authhandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
)

func (apicfg *HandlersAuthConfig) HandlerSignIn(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)

	params, err := auth.DecodeAndValidate[parameters](w, r)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"invalid request body",
			"Invalid signin payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := apicfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"get user failed",
			"Error getting user from email",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}

	err = apicfg.AuthHelper.CheckPasswordHash(params.Password, user.Password.String)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"check password failed",
			"Error checking password",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"uuid parse failed",
			"Error parsing uuid",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(handlers.AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(handlers.RefreshTokenTTL)

	accessToken, refreshToken, err := apicfg.AuthHelper.GenerateTokens(userID.String(), accessTokenExpiresAt)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"generate tokens failed",
			"Error generating tokens",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdateUserStatusByID(r.Context(), database.UpdateUserStatusByIDParams{
		ID:        user.ID,
		Provider:  "local",
		UpdatedAt: timeNow,
	})
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"update user status failed",
			"Error updating user status in database",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	err = apicfg.AuthHelper.StoreRefreshTokenInRedis(r, userID.String(), refreshToken, "local", refreshTokenExpiresAt.Sub(timeNow))
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"store refresh token failed",
			"Error saving refresh token to Redis",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signin-local",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	auth.SetTokensAsCookies(w, accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userID.String())
	apicfg.LogHandlerSuccess(ctxWithUserID, "signin-local", "Local signin success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Signin successful",
	})
}
