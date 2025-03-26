package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
)

func (apicfg *HandlersConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Refresh token required")
		return
	}

	refreshToken := cookie.Value

	userID, err := apicfg.Auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Println("Invalid refresh token:", err)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	err = apicfg.RedisClient.Del(r.Context(), "refresh_token:"+userID.String()).Err()
	if err != nil {
		log.Println("Error deleting refresh token from Redis:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	timeNow := time.Now().UTC()
	newKeyExpiredAt := timeNow.Add(-1 * time.Hour)

	err = apicfg.DB.UpdateUserStatusByID(r.Context(), database.UpdateUserStatusByIDParams{
		UpdatedAt: timeNow,
		Provider:  "local",
		ID:        userID.String(),
	})
	if err != nil {
		log.Println("Error update signout status to database: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  newKeyExpiredAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  newKeyExpiredAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Sign out successful",
	})
}
