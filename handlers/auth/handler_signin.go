package authhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerSignIn handles user authentication requests
// It validates the signin payload, authenticates the user,
// generates authentication tokens, merges guest cart if needed,
// and returns a success response
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
