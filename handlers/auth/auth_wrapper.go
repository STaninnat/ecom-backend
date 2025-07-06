package authhandlers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/handlers/cart"
	"github.com/STaninnat/ecom-backend/middlewares"
)

// HandlersAuthConfig contains the configuration for auth handlers
// It embeds the base handlers config and cart config, and provides
// access to the auth service for business logic operations
type HandlersAuthConfig struct {
	*handlers.HandlersConfig
	*cart.HandlersCartConfig
	authService AuthService
	authMutex   sync.RWMutex
}

// InitAuthService initializes the auth service with the current configuration
// This method should be called during application startup
func (cfg *HandlersAuthConfig) InitAuthService() error {
	// Validate that the embedded config is not nil
	if cfg.HandlersConfig == nil {
		return errors.New("handlers config not initialized")
	}
	if cfg.APIConfig == nil {
		return errors.New("API config not initialized")
	}
	// Validate required dependencies
	if cfg.DB == nil {
		return errors.New("database not initialized")
	}
	if cfg.Auth == nil {
		return errors.New("auth config not initialized")
	}
	if cfg.RedisClient == nil {
		return errors.New("redis client not initialized")
	}

	cfg.authMutex.Lock()
	defer cfg.authMutex.Unlock()

	cfg.authService = NewAuthService(
		cfg.DB,
		cfg.DBConn,
		cfg.Auth,
		cfg.RedisClient,
		cfg.OAuth.Google,
	)

	return nil
}

// GetAuthService returns the auth service instance, initializing it if necessary
// This method is thread-safe and will initialize the service on first access
func (cfg *HandlersAuthConfig) GetAuthService() AuthService {
	cfg.authMutex.RLock()
	if cfg.authService != nil {
		defer cfg.authMutex.RUnlock()
		return cfg.authService
	}
	cfg.authMutex.RUnlock()

	// Need to initialize, acquire write lock
	cfg.authMutex.Lock()
	defer cfg.authMutex.Unlock()

	// Double-check pattern in case another goroutine initialized it
	if cfg.authService == nil {
		// Validate that the embedded config is not nil before accessing its fields
		if cfg.HandlersConfig == nil || cfg.APIConfig == nil || cfg.DB == nil {
			// Return a default service that will fail gracefully when used
			cfg.authService = NewAuthService(nil, nil, nil, nil, nil)
		} else {
			cfg.authService = NewAuthService(
				cfg.DB,
				cfg.DBConn,
				cfg.Auth,
				cfg.RedisClient,
				cfg.OAuth.Google,
			)
		}
	}

	return cfg.authService
}

// handleAuthError handles authentication-specific errors with proper logging and responses
// It categorizes errors and provides appropriate HTTP status codes and messages
func (cfg *HandlersAuthConfig) handleAuthError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if authErr, ok := err.(*AuthError); ok {
		switch authErr.Code {
		case "name_exists", "email_exists", "user_not_found", "invalid_password":
			cfg.LogHandlerError(ctx, operation, authErr.Code, authErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, authErr.Message)
		case "database_error", "transaction_error", "create_user_error", "hash_error", "token_generation_error", "redis_error", "commit_error", "update_user_error", "uuid_error":
			cfg.LogHandlerError(ctx, operation, authErr.Code, authErr.Message, ip, userAgent, authErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "invalid_state", "token_exchange_error", "google_api_error", "no_refresh_token", "google_token_error":
			cfg.LogHandlerError(ctx, operation, authErr.Code, authErr.Message, ip, userAgent, authErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, authErr.Message)
		default:
			cfg.LogHandlerError(ctx, operation, "internal_error", authErr.Message, ip, userAgent, authErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
