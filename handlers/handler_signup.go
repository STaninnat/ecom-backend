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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
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

	if !auth.IsValidateEmailFormat(params.Email) {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	// ------------------------------------------------------------

	nameExists, err := apicfg.DB.CheckUserExistsByName(r.Context(), params.Name)
	if err != nil {
		log.Println("Error checking username exist: ", err)
		return
	}
	if nameExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this name already exists")
		return
	}

	emailExists, err := apicfg.DB.CheckUserExistsByEmail(r.Context(), params.Email)
	if err != nil {
		log.Println("Error checking email exists: ", err)
		return
	}

	if emailExists {
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this email already exists")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Println("Error couldn't hash password: ", err)
		return
	}

	userID := uuid.New()

	err = apicfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        userID.String(),
		Name:      params.Name,
		Email:     params.Email,
		Password:  sql.NullString{String: hashedPassword, Valid: true},
		Provider:  "local",
		CreatedAt: time.Now().Local(),
		UpdatedAt: time.Now().Local(),
	})
	if err != nil {
		log.Println("Error couldn't create user in database: ", err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		return
	}

	accessTokenExpiresAt := time.Now().Local().Add(30 * time.Minute)
	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, apicfg.JWTSecret, accessTokenExpiresAt)
	if err != nil {
		log.Println("Error couldn't generate access token: ", err)
		return
	}

	refreshTokenExpiresAt := time.Now().Local().Add(30 * 24 * time.Hour)
	refreshToken, err := apicfg.Auth.GenerateRefreshToken()
	if err != nil {
		log.Println("Error couldn't generate refresh token: ", err)
		return
	}

	err = apicfg.DB.CreateUserSession(r.Context(), database.CreateUserSessionParams{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		Token:     refreshToken,
		ExpiresAt: refreshTokenExpiresAt,
		CreatedAt: time.Now().Local(),
		UpdatedAt: time.Now().Local(),
	})
	if err != nil {
		log.Println("Error saving refresh token: ", err)
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
