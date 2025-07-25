// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
package authhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_oauth.go: Provides handlers for initiating and processing Google OAuth signin.

// HandlerGoogleSignIn initiates the Google OAuth signin process.
// @Summary      Google OAuth signin
// @Description  Redirects to Google OAuth for authentication
// @Tags         auth
// @Produce      json
// @Success      302  {string}  string  "Redirect"
// @Failure      500  {object}  map[string]string
// @Router       /v1/auth/google/signin [get]
func (cfg *HandlersAuthConfig) HandlerGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	// Generate state and auth URL
	state := auth.GenerateState()
	authURL, err := cfg.GetAuthService().GenerateGoogleAuthURL(state)
	if err != nil {
		cfg.Logger.LogHandlerError(
			r.Context(),
			"signin-google",
			"auth_url_generation_failed",
			"Error generating Google auth URL",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to initiate Google signin")
		return
	}

	// Redirect to Google
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandlerGoogleCallback handles the Google OAuth callback.
// @Summary      Google OAuth callback
// @Description  Handles Google OAuth callback and authenticates user
// @Tags         auth
// @Produce      json
// @Success      201  {object}  handlers.HandlerResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/auth/google/callback [get]
func (cfg *HandlersAuthConfig) HandlerGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get parameters from URL
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"callback-google",
			"missing_parameters",
			"Missing state or code parameter",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().HandleGoogleAuth(ctx, code, state)
	if err != nil {
		cfg.handleAuthError(w, r, err, "callback-google", ip, userAgent)
		return
	}

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, result.UserID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "callback-google", "Google signin success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Google signin successful",
	})
}
