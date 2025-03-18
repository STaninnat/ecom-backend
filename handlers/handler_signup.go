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

type parameters struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (apicfg *HandlersConfig) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Println("Decode error: ", err)
		return
	}

	// --------------------move to frontend-------------------
	if params.Name == "" || params.Email == "" || params.Password == "" {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if !auth.IsValidUserNameFormat(params.Name) {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid username format")
		return
	}

	if !auth.IsValidEmailFormat(params.Email) {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	// ------------------------------------------------------------

	if nameExists, err := apicfg.DB.CheckUserExistsByName(r.Context(), params.Name); err != nil || nameExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this name already exists")
		return
	}

	if emailExists, err := apicfg.DB.CheckUserExistsByEmail(r.Context(), params.Email); err != nil || emailExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this email already exists")
		return
	}

	userID := uuid.New()
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Println("Error couldn't hash password: ", err)
		return
	}

	timeNow := time.Now().Local()
	accessTokenExpiresAt := time.Now().Local().Add(30 * time.Minute)
	refreshTokenExpiresAt := time.Now().Local().Add(7 * 24 * time.Hour)

	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, apicfg.JWTSecret, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error generate access token: ", err)
		return
	}

	refreshToken, err := apicfg.Auth.GenerateRefreshToken()
	if err != nil {
		log.Println("Error generate refresh token: ", err)
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
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		return
	}

	err = apicfg.RedisClient.Set(r.Context(), "refresh_token:"+userID.String(), refreshToken, refreshTokenExpiresAt.Sub(time.Now().Local())).Err()
	if err != nil {
		log.Println("Error saving refresh token to Redis: ", err)
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
