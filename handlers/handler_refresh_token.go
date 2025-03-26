package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/middlewares"
)

func (apicfg *HandlersConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Refresh token is required")
		return
	}
	refreshToken := cookie.Value

	userID, err := apicfg.Auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Println("Invalid refresh token:", err)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	storedToken, err := apicfg.RedisClient.Get(r.Context(), "refresh_token:"+userID.String()).Result()
	if err != nil || storedToken != refreshToken {
		log.Println("Refresh token mismatch:", err)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	err = apicfg.RedisClient.Del(r.Context(), "refresh_token:"+userID.String()).Err()
	if err != nil {
		log.Println("Error deleting old refresh token:", err)
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, apicfg.JWTSecret, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error generating access token:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	newRefreshToken, err := apicfg.Auth.GenerateRefreshToken(userID)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	err = apicfg.RedisClient.Set(r.Context(), "refresh_token:"+userID.String(), newRefreshToken, refreshTokenExpiresAt.Sub(timeNow)).Err()
	if err != nil {
		log.Println("Error saving refresh token to Redis:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Token refreshed successful",
	})
}
