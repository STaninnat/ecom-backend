package authhandlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
)

func (apicfg *HandlersAuthConfig) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)

	params, err := auth.DecodeAndValidate[parameters](w, r)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"invalid request",
			"Invalid signup payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	nameExists, err := apicfg.DB.CheckUserExistsByName(r.Context(), params.Name)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"check name failed",
			"Error checking name existence",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if nameExists {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"name exists",
			"Duplicate name",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this name already exists")
		return
	}

	emailExists, err := apicfg.DB.CheckUserExistsByEmail(r.Context(), params.Email)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"check email failed",
			"Error checking email existence",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if emailExists {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"email exists",
			"Duplicate email",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "An account with this email already exists")
		return
	}

	userID := uuid.New()
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"hash password failed",
			"Error hashing password",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	timeNow := time.Now().UTC()

	tx, err := apicfg.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
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
		Role:       "user",
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
	})
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"create user failed",
			"Error creating user in database",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		return
	}

	accessTokenExpiresAt := timeNow.Add(handlers.AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(handlers.RefreshTokenTTL)

	accessToken, refreshToken, err := apicfg.AuthHelper.GenerateTokens(userID.String(), accessTokenExpiresAt)
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"generate tokens failed",
			"Error generating tokens",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	err = apicfg.AuthHelper.StoreRefreshTokenInRedis(r, userID.String(), refreshToken, "local", refreshTokenExpiresAt.Sub(timeNow))
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"store refresh token failed",
			"Error saving refresh token to Redis",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"signup-local",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	auth.SetTokensAsCookies(w, accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userID.String())
	apicfg.LogHandlerSuccess(ctxWithUserID, "signup-local", "Local signup success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Signup successful",
	})
}
