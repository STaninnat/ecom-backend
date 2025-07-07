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

const (
	GoogleProvider = "google"
)

// HandlerSignOut handles user logout requests
// It validates the user's authentication, clears their session,
// revokes tokens, clears cookies, and handles OAuth provider-specific logout
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

	// Log success and respond
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID.String())
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "sign_out", "Sign out success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Sign out successful",
	})
}
