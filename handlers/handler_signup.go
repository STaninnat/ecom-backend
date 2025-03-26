package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/google/uuid"
)

func (apicfg *HandlersConfig) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params, valid := auth.DecodeAndValidate[parameters](w, r)
	if !valid {
		return
	}

	nameExists, err := apicfg.DB.CheckUserExistsByName(r.Context(), params.Name)
	if err != nil {
		log.Println("Error checking name existence:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if nameExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this name already exists")
		return
	}

	emailExists, err := apicfg.DB.CheckUserExistsByEmail(r.Context(), params.Email)
	if err != nil {
		log.Println("Error checking email existence:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if emailExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this email already exists")
		return
	}

	userID := uuid.New()
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Println("Error couldn't hash password: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	timeNow := time.Now().UTC()

	tx, err := apicfg.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}

	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.CreateUser(r.Context(), database.CreateUserParams{
		ID:         userID.String(),
		Name:       params.Name,
		Email:      params.Email,
		Password:   sql.NullString{String: hashedPassword, Valid: true},
		Provider:   "local",
		ProviderID: sql.NullString{},
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
	})
	if err != nil {
		log.Println("Error creating user in database:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		return
	}

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

	middlewares.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Signup successful",
	})
}
