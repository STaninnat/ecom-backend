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

// HandlerSignUp handles user registration requests.
// Validates the signup payload, creates a new user account, generates tokens, merges guest cart if needed, and returns a success response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
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
// Validates the signin payload, authenticates the user, generates tokens, merges guest cart if needed, and returns a success response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
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
// Validates authentication, clears session, revokes tokens, clears cookies, and handles OAuth provider-specific logout.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
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
