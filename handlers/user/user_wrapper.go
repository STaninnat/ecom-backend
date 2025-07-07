package userhandlers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/STaninnat/ecom-backend/handlers"
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
	if userErr, ok := err.(*UserError); ok {
		switch userErr.Code {
		case "transaction_error", "update_failed", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, userErr.Code, userErr.Message, ip, userAgent, userErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", userErr.Message, ip, userAgent, userErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
