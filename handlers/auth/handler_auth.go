// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
package authhandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_auth.go: Provides HTTP handlers for user signup, signin, and signout with token management.

// SignupRequest represents the payload for user signup.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SigninRequest represents the payload for user signin.
type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandlerSignUp handles user registration requests.
// @Summary      User signup
// @Description  Registers a new user and returns tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        signup  body  SignupRequest  true  "Signup payload"
// @Success      201  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/auth/signup [post]
func (cfg *HandlersAuthConfig) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Parse and validate request
	params, err := auth.DecodeAndValidate[struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"signup-local",
			"invalid_request",
			"Invalid signup payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().SignUp(ctx, SignUpParams{
		Name:     params.Name,
		Email:    params.Email,
		Password: params.Password,
	})

	if err != nil {
		cfg.handleAuthError(w, r, err, "signup-local", ip, userAgent)
		return
	}

	// Merge cart if needed
	cfg.MergeCart(ctx, r, result.UserID)

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, result.UserID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "signup-local", "Local signup success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Signup successful",
	})
}

// HandlerSignIn handles user authentication requests.
// @Summary      User signin
// @Description  Authenticates a user and returns tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        signin  body  SigninRequest  true  "Signin payload"
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/auth/signin [post]
func (cfg *HandlersAuthConfig) HandlerSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Parse and validate request
	params, err := auth.DecodeAndValidate[struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"signin-local",
			"invalid_request",
			"Invalid signin payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().SignIn(ctx, SignInParams{
		Email:    params.Email,
		Password: params.Password,
	})

	if err != nil {
		cfg.handleAuthError(w, r, err, "signin-local", ip, userAgent)
		return
	}

	// Merge cart if needed
	cfg.MergeCart(ctx, r, result.UserID)

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, result.UserID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "signin-local", "Local signin success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Signin successful",
	})
}

// HandlerSignOut handles user logout requests.
// @Summary      User signout
// @Description  Logs out the user and revokes tokens
// @Tags         auth
// @Produce      json
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      401  {object}  map[string]string
// @Router       /v1/auth/signout [post]
func (cfg *HandlersAuthConfig) HandlerSignOut(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"sign_out",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	err = cfg.GetAuthService().SignOut(ctx, userID.String(), storedData.Provider)
	if err != nil {
		cfg.handleAuthError(w, r, err, "sign_out", ip, userAgent)
		return
	}

	// Clear cookies
	timeNow := time.Now().UTC()
	expiredTime := timeNow.Add(-1 * time.Hour)
	auth.SetTokensAsCookies(w, "", "", expiredTime, expiredTime)

	// Handle Google revoke if needed
	if storedData.Provider == GoogleProvider {
		googleRevokeURL := "https://accounts.google.com/o/oauth2/revoke?token=" + storedData.Token
		http.Redirect(w, r, googleRevokeURL, http.StatusFound)
		return
	}

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID.String())
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "sign_out", "Sign out success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Sign out successful",
	})
}
