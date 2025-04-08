package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/middlewares"
)

func (apicfg *HandlersConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	userID, storedData, err := apicfg.AuthHelper.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if storedData.Provider == "local" {
		err = apicfg.RedisClient.Del(r.Context(), "refresh_token:"+userID.String()).Err()
		if err != nil {
			log.Println("Error deleting refresh token from Redis:", err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
			return
		}
	}

	timeNow := time.Now().UTC()
	newKeyExpiredAt := timeNow.Add(-1 * time.Hour)

	auth.SetTokensAsCookies(w, "", "", newKeyExpiredAt, newKeyExpiredAt)

	if storedData.Provider == "google" {
		googleRevokeURL := "https://accounts.google.com/o/oauth2/revoke?token=" + storedData.Token
		http.Redirect(w, r, googleRevokeURL, http.StatusFound)
		return
	}

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Sign out successful",
	})
}
