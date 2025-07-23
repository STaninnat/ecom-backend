// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
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

// user_wrapper.go: Provides thread-safe user service setup, error handling, user extraction middleware, and auth handlers.

// HandlersUserConfig contains configuration and dependencies for user handlers.
// Embeds Config, provides logger, userService, and thread safety.
// Manages the lifecycle of user service instances with proper synchronization.
type HandlersUserConfig struct {
	Config      *handlers.Config       // for DB, etc.
	Logger      handlers.HandlerLogger // for logging
	userService UserService
	userMutex   sync.RWMutex
}

// InitUserService initializes the user service with the current configuration.
// Validates that handlers config and database are initialized before creating the service.
// Thread-safe operation using mutex for concurrent access.
// Returns:
//   - error: nil on success, error if handlers config or database is not initialized
func (cfg *HandlersUserConfig) InitUserService() error {
	if cfg.Config == nil {
		return errors.New("handlers config not initialized")
	}
	if cfg.Config.DB == nil {
		return errors.New("database not initialized")
	}
	cfg.userMutex.Lock()
	defer cfg.userMutex.Unlock()
	cfg.userService = NewUserService(cfg.Config.DB, cfg.Config.DBConn)
	return nil
}

// GetUserService returns the user service instance, initializing it if necessary.
// Uses double-checked locking pattern for thread-safe lazy initialization.
// Creates service with available dependencies or fallback to nil dependencies.
// Returns:
//   - UserService: the current user service instance
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
		if cfg.Config == nil || cfg.Config.DB == nil {
			cfg.userService = NewUserService(nil, nil)
		} else {
			cfg.userService = NewUserService(cfg.Config.DB, cfg.Config.DBConn)
		}
	}
	return cfg.userService
}

// handleUserError handles user-specific errors with proper logging and responses.
// Maps AppError codes to corresponding HTTP status codes and user-friendly messages.
// Parameters:
//   - w: http.ResponseWriter for sending the error response
//   - r: *http.Request containing the request context
//   - err: error to handle
//   - operation: string describing the operation that failed
//   - ip: string client IP address
//   - userAgent: string client user agent
func (cfg *HandlersUserConfig) handleUserError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	var appErr *handlers.AppError
	if errors.As(err, &appErr) {
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
// Extracts JWT token from Authorization header, validates it, and fetches user from database.
// Sets user in request context for downstream handlers to access.
// Parameters:
//   - next: http.Handler to call after user extraction
//
// Returns:
//   - http.Handler: middleware that processes requests with user context
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

// extractUserFromRequest is a placeholder for your actual user extraction logic.
// Extracts JWT from Authorization header, validates token, and fetches user from database.
// Parameters:
//   - r: *http.Request containing the Authorization header
//
// Returns:
//   - database.User: the extracted user
//   - error: nil on success, error on failure
func (cfg *HandlersUserConfig) extractUserFromRequest(r *http.Request) (user database.User, err error) {
	// Extract JWT from Authorization header
	header := r.Header.Get("Authorization")
	if header == "" || len(header) < 8 || header[:7] != "Bearer " {
		return database.User{}, errors.New("missing or invalid Authorization header")
	}
	token := header[7:]

	// Validate JWT and extract claims
	claims, err := cfg.Config.Auth.ValidateAccessToken(token, cfg.Config.JWTSecret)
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

// AuthHandlerGetUser wraps HandlerGetUser with user context for authenticated routes.
// Sets the authenticated user in request context before calling the handler.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersUserConfig) AuthHandlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerGetUser(w, r.WithContext(ctx))
}

// AuthHandlerUpdateUser updates user information for the authenticated user.
func (cfg *HandlersUserConfig) AuthHandlerUpdateUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerUpdateUser(w, r.WithContext(ctx))
}

// AuthHandlerPromoteUserToAdmin promotes a user to admin status for the authenticated user.
func (cfg *HandlersUserConfig) AuthHandlerPromoteUserToAdmin(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerPromoteUserToAdmin(w, r.WithContext(ctx))
}
