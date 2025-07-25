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

// handler_refresh_token.go: Handles the refresh token flow by validating and issuing new tokens.

// HandlerRefreshToken handles token refresh requests.
// @Summary      Refresh token
// @Description  Refreshes access and refresh tokens
// @Tags         auth
// @Produce      json
// @Success      200  {object}  handlers.HandlerResponse
// @Failure      401  {object}  map[string]string
// @Router       /v1/auth/refresh [post]
func (cfg *HandlersAuthConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get user info from token
	userID, storedData, err := cfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"refresh_token",
			"invalid_token",
			"Error validating authentication token",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().RefreshToken(ctx, userID.String(), storedData.Provider, storedData.Token)
	if err != nil {
		cfg.handleAuthError(w, r, err, "refresh_token", ip, userAgent)
		return
	}

	// Set new cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID.String())
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "refresh_token", "Refresh token success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Token refreshed successfully",
	})
}
