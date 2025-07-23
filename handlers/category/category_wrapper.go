// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
)

// category_wrapper.go: Provides configuration, initialization, and error handling for category-related operations.

// HandlersCategoryConfig contains the configuration for category handlers.
// Manages the category service lifecycle and provides thread-safe access to the service instance.
type HandlersCategoryConfig struct {
	*handlers.Config
	Logger          handlers.HandlerLogger
	categoryService CategoryService
	categoryMutex   sync.RWMutex
}

// InitCategoryService initializes the category service with the current configuration.
// Validates required dependencies and sets up the service. Returns an error if any dependency is missing.
func (cfg *HandlersCategoryConfig) InitCategoryService() error {
	// Validate that the embedded config is not nil
	if cfg.Config == nil {
		return errors.New("handlers config not initialized")
	}
	if cfg.APIConfig == nil {
		return errors.New("API config not initialized")
	}
	// Validate required dependencies
	if cfg.DB == nil {
		return errors.New("database not initialized")
	}
	if cfg.DBConn == nil {
		return errors.New("database connection not initialized")
	}

	cfg.categoryMutex.Lock()
	defer cfg.categoryMutex.Unlock()

	cfg.categoryService = NewCategoryService(cfg.DB, cfg.DBConn)

	// Set Logger if not already set
	if cfg.Logger == nil {
		cfg.Logger = cfg.Config // Config implements HandlerLogger
	}

	return nil
}

// GetCategoryService returns the category service instance, initializing it if necessary.
// Uses a double-checked locking pattern for thread-safe lazy initialization. If dependencies are missing, creates a service with nil dependencies.
func (cfg *HandlersCategoryConfig) GetCategoryService() CategoryService {
	cfg.categoryMutex.RLock()
	if cfg.categoryService != nil {
		defer cfg.categoryMutex.RUnlock()
		return cfg.categoryService
	}
	cfg.categoryMutex.RUnlock()

	// Need to initialize, acquire write lock
	cfg.categoryMutex.Lock()
	defer cfg.categoryMutex.Unlock()

	// Double-check pattern in case another goroutine initialized it
	if cfg.categoryService == nil {
		// Validate that the embedded config is not nil before accessing its fields
		if cfg.Config == nil || cfg.APIConfig == nil || cfg.DB == nil || cfg.DBConn == nil {
			// Return a default service that will fail gracefully when used
			cfg.categoryService = NewCategoryService(nil, nil)
		} else {
			cfg.categoryService = NewCategoryService(cfg.DB, cfg.DBConn)
		}
	}

	return cfg.categoryService
}

// handleCategoryError handles category-specific errors with proper logging and responses.
// Categorizes errors and provides appropriate HTTP status codes and messages. All errors are logged with context information for debugging.
func (cfg *HandlersCategoryConfig) handleCategoryError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	var appErr *handlers.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "invalid_request":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "database_error", "transaction_error", "create_category_error", "update_category_error", "delete_category_error", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// CategoryWithIDRequest represents a request containing a category ID and optional name and description.
type CategoryWithIDRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
