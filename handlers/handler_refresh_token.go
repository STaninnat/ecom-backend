package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/middlewares"
	"golang.org/x/oauth2"
)

func (apicfg *HandlersConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	userID, storedData, err := apicfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		return
	}

	if storedData.Provider == "google" {
		refreshToken := storedData.Token
		log.Println("Using refresh token:", refreshToken)

		newToken, err := apicfg.RefreshGoogleAccessToken(refreshToken)
		if err != nil {
			fmt.Println("Failed to refresh Google token: ", err)
			middlewares.RespondWithError(w, http.StatusUnauthorized, "Failed to refresh Google token")
			return
		}

		auth.SetTokensAsCookies(w, newToken.AccessToken, refreshToken, newToken.Expiry, time.Now().Add(7*24*time.Hour))
		middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Token refreshed successful",
		})
		return
	}

	err = apicfg.RedisClient.Del(r.Context(), "refresh_token:"+userID.String()).Err()
	if err != nil {
		log.Println("Error deleting old refresh token:", err)
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	accessToken, newRefreshToken, err := apicfg.Auth.GenerateTokens(userID.String(), accessTokenExpiresAt)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	err = apicfg.Auth.StoreRefreshTokenInRedis(r, userID.String(), newRefreshToken, "local", refreshTokenExpiresAt.Sub(timeNow))
	if err != nil {
		log.Println("Error saving refresh token to Redis: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	auth.SetTokensAsCookies(w, accessToken, newRefreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Token refreshed successful",
	})
}

func (apicfg *HandlersConfig) RefreshGoogleAccessToken(refreshToken string) (*oauth2.Token, error) {
	tokenSource := apicfg.OAuth.Google.TokenSource(context.Background(), &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}
