package userhandlers

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
)

type HandlersUserConfig struct {
	HandlersConfig *handlers.HandlersConfig // for DB, etc.
	Logger         handlers.HandlerLogger   // for logging
	userService    UserService
	userMutex      sync.RWMutex
}

// InitUserService initializes the user service with the current configuration
func (cfg *HandlersUserConfig) InitUserService() error {
	if cfg.HandlersConfig == nil {
		return errors.New("handlers config not initialized")
	}
	if cfg.HandlersConfig.DB == nil {
		return errors.New("database not initialized")
	}
	cfg.userMutex.Lock()
	defer cfg.userMutex.Unlock()
	cfg.userService = NewUserService(cfg.HandlersConfig.DB, cfg.HandlersConfig.DBConn)
	return nil
}

// GetUserService returns the user service instance, initializing it if necessary
func (cfg *HandlersUserConfig) GetUserService() UserService {
	cfg.userMutex.RLock()
	if cfg.userService != nil {
		defer cfg.userMutex.RUnlock()
		return cfg.userService
	}
	cfg.userMutex.RUnlock()
	cfg.userMutex.Lock()
	defer cfg.userMutex.Unlock()
	if cfg.userService == nil {
		if cfg.HandlersConfig == nil || cfg.HandlersConfig.DB == nil {
			cfg.userService = NewUserService(nil, nil)
		} else {
			cfg.userService = NewUserService(cfg.HandlersConfig.DB, cfg.HandlersConfig.DBConn)
		}
	}
	return cfg.userService
}

// handleUserError handles user-specific errors with proper logging and responses
func (cfg *HandlersUserConfig) handleUserError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if appErr, ok := err.(*handlers.AppError); ok {
		switch appErr.Code {
		case "transaction_error", "update_failed", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "user_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "invalid_request":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// UserExtractionMiddleware extracts the user from the request and sets it in the context using contextKeyUser.
func (cfg *HandlersUserConfig) UserExtractionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Example: extract user from a JWT in the Authorization header (pseudo-code)
		// You should replace this with your actual authentication logic
		user, err := cfg.extractUserFromRequest(r)
		if err != nil {
			// Optionally, you can handle unauthorized here or let the handler do it
			next.ServeHTTP(w, r) // Pass through without user in context
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUserFromRequest is a placeholder for your actual user extraction logic
func (cfg *HandlersUserConfig) extractUserFromRequest(r *http.Request) (user database.User, err error) {
	// Extract JWT from Authorization header
	header := r.Header.Get("Authorization")
	if header == "" || len(header) < 8 || header[:7] != "Bearer " {
		return database.User{}, errors.New("missing or invalid Authorization header")
	}
	token := header[7:]

	// Validate JWT and extract claims
	claims, err := cfg.HandlersConfig.Auth.ValidateAccessToken(token, cfg.HandlersConfig.JWTSecret)
	if err != nil {
		return database.User{}, errors.New("invalid token")
	}

	// Fetch user from DB
	user, err = cfg.GetUserService().GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		return database.User{}, errors.New("user not found")
	}

	return user, nil
}
