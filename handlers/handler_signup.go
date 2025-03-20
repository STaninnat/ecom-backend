package handlers

import (
	"database/sql"
	"encoding/json"
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

	defer r.Body.Close()

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Println("Decode error: ", err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if params.Name == "" || params.Email == "" || params.Password == "" {
		log.Println("Invalid request format")
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if nameExists, err := apicfg.DB.CheckUserExistsByName(r.Context(), params.Name); err != nil {
		log.Println("Error checking name existence:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	} else if nameExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this name already exists")
		return
	}

	if emailExists, err := apicfg.DB.CheckUserExistsByEmail(r.Context(), params.Email); err != nil {
		log.Println("Error checking email existence:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	} else if emailExists {
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
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, apicfg.JWTSecret, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error generate access token: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := apicfg.Auth.GenerateRefreshToken()
	if err != nil {
		log.Println("Error generate refresh token: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	err = apicfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        userID.String(),
		Name:      params.Name,
		Email:     params.Email,
		Password:  sql.NullString{String: hashedPassword, Valid: true},
		Provider:  "local",
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	})
	if err != nil {
		log.Println("Error creating user in database:", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		return
	}

	err = apicfg.RedisClient.Set(r.Context(), "refresh_token:"+userID.String(), refreshToken, refreshTokenExpiresAt.Sub(timeNow)).Err()
	if err != nil {
		log.Println("Error saving refresh token to Redis: ", err)
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
		Value:    refreshToken,
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	userResp := map[string]string{
		"message": "Signup successful",
	}

	middlewares.RespondWithJSON(w, http.StatusCreated, userResp)
}
