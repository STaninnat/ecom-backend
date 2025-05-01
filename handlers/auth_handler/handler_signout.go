package auth_handler

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersAuthConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	userID, storedData, err := apicfg.AuthHelper.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "signout", "validate cookie failed", "Error validating cookie", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = apicfg.RedisClient.Del(r.Context(), auth.RedisRefreshTokenPrefix+userID.String()).Err()
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "signout", "delete token failed", "Error deleting refresh token from Redis", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove refresh token from Redis")
		return
	}

	timeNow := time.Now().UTC()
	newKeyExpiredAt := timeNow.Add(-1 * time.Hour)

	auth.SetTokensAsCookies(w, "", "", newKeyExpiredAt, newKeyExpiredAt)

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userID.String())

	if storedData.Provider == "google" {
		apicfg.LogHandlerSuccess(ctxWithUserID, "signout", "Sign out success", ip, userAgent)

		googleRevokeURL := "https://accounts.google.com/o/oauth2/revoke?token=" + storedData.Token
		http.Redirect(w, r, googleRevokeURL, http.StatusFound)
		return
	}

	apicfg.LogHandlerSuccess(ctxWithUserID, "signout", "Sign out success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Sign out successful",
	})
}
