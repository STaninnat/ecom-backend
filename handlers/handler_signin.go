package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/google/uuid"
)

func (apicfg *HandlersConfig) HandlerSignIn(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params, valid := auth.DecodeAndValidate[parameters](w, r)
	if !valid {
		return
	}

	user, err := apicfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}

	if !auth.CheckPasswordHash(params.Password, user.Password.String) {
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		log.Printf("Error parsing user id: %v\n", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, apicfg.JWTSecret, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error generate access token: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := apicfg.Auth.GenerateRefreshToken(userID)
	if err != nil {
		log.Println("Error generate refresh token: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdateUserStatusByID(r.Context(), database.UpdateUserStatusByIDParams{
		UpdatedAt: timeNow,
		Provider:  "local",
		ID:        user.ID,
	})
	if err != nil {
		log.Println("Error update signin status to database: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	err = apicfg.RedisClient.Set(r.Context(), "refresh_token:"+userID.String(), refreshToken, refreshTokenExpiresAt.Sub(timeNow)).Err()
	if err != nil {
		log.Println("Error saving refresh token to Redis: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
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
		Value:    refreshToken,
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Signin successful",
	})
}
