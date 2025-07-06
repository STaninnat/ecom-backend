package authhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerSignUp handles user registration requests
// It validates the signup payload, creates a new user account,
// generates authentication tokens, merges guest cart if needed,
// and returns a success response
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
		cfg.LogHandlerError(
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
	cfg.LogHandlerSuccess(ctxWithUserID, "signup-local", "Local signup success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Signup successful",
	})
}
