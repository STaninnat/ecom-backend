// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
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
	userhandlers.HandleErrorWithCodeMap(cfg.Logger, w, r, err, operation, ip, userAgent, categoryErrorCodeMap, http.StatusInternalServerError, "Internal server error")
}

// HandleCategoryRequest is a shared helper for create/update category handlers (production and test).
func HandleCategoryRequest[
	S any, // Service type (CategoryService or mock)
	L handlers.HandlerLogger, // Logger type
](
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
	logger L,
	getCategoryService func() S,
	handleCategoryError func(http.ResponseWriter, *http.Request, error, string, string, string),
	action string,
	serviceFunc func(context.Context, S, CategoryRequest) (string, error), // for create: returns id, for update: returns empty string
	successMessage string,
	successStatus int,
) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.LogHandlerError(
			ctx,
			action,
			"invalid_request_body",
			"Failed to parse request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	categoryService := getCategoryService()
	_, err := serviceFunc(ctx, categoryService, params)
	if err != nil {
		handleCategoryError(w, r, err, action, ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	logger.LogHandlerSuccess(ctxWithUserID, action, successMessage, ip, userAgent)
	middlewares.RespondWithJSON(w, successStatus, handlers.HandlerResponse{
		Message: successMessage,
	})
}

// CategoryWithIDRequest represents a request containing a category ID and optional name and description.
type CategoryWithIDRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Extract shared codeMap for category error handling
type categoryErrorCodeMapType map[string]userhandlers.ErrorResponseConfig

var categoryErrorCodeMap = categoryErrorCodeMapType{
	"invalid_request":       {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
	"database_error":        {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
	"transaction_error":     {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
	"create_category_error": {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
	"update_category_error": {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
	"delete_category_error": {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
	"commit_error":          {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
}

// SharedHandleCategoryError is a shared error handler for category operations (production and test).
func SharedHandleCategoryError(
	logger handlers.HandlerLogger,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	operation, ip, userAgent string,
) {
	userhandlers.HandleErrorWithCodeMap(logger, w, r, err, operation, ip, userAgent, categoryErrorCodeMap, http.StatusInternalServerError, "Internal server error")
}

// HandleCategoryDelete is a shared helper for delete category handlers (production and test).
func HandleCategoryDelete[
	S any, // Service type (CategoryService or mock)
	L handlers.HandlerLogger, // Logger type
](
	w http.ResponseWriter,
	r *http.Request,
	user database.User,
	logger L,
	getCategoryService func() S,
	sharedHandleCategoryError func(handlers.HandlerLogger, http.ResponseWriter, *http.Request, error, string, string, string),
) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		logger.LogHandlerError(
			ctx,
			"delete_category",
			"missing_category_id",
			"Category ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	categoryService := getCategoryService()
	err := any(categoryService).(interface {
		DeleteCategory(context.Context, string) error
	}).DeleteCategory(ctx, categoryID)
	if err != nil {
		sharedHandleCategoryError(logger, w, r, err, "delete_category", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	logger.LogHandlerSuccess(ctxWithUserID, "delete_category", "Category deleted successfully", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Category deleted successfully",
	})
}
