package producthandlers

import (
	"database/sql"
	"errors"
	"net/http"
	"sync"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlersProductConfig holds the configuration and dependencies for product handlers.
// It manages the product service lifecycle and provides thread-safe access to the service instance.
type HandlersProductConfig struct {
	DB             *database.Queries
	DBConn         *sql.DB
	Logger         handlers.HandlerLogger // for logging
	productService ProductService
	productMutex   sync.RWMutex
}

// InitProductService initializes the product service with the current configuration.
// It validates that both DB and DBConn are set before creating the service.
// Returns an error if either dependency is missing.
func (cfg *HandlersProductConfig) InitProductService() error {
	if cfg.DB == nil {
		return errors.New("database not initialized")
	}
	if cfg.DBConn == nil {
		return errors.New("database connection not initialized")
	}
	cfg.productMutex.Lock()
	defer cfg.productMutex.Unlock()
	cfg.productService = NewProductService(cfg.DB, cfg.DBConn)
	return nil
}

// GetProductService returns the product service instance, initializing it if necessary.
// It uses a double-checked locking pattern for thread-safe lazy initialization.
// If dependencies are missing, it creates a service with nil dependencies.
func (cfg *HandlersProductConfig) GetProductService() ProductService {
	cfg.productMutex.RLock()
	if cfg.productService != nil {
		defer cfg.productMutex.RUnlock()
		return cfg.productService
	}
	cfg.productMutex.RUnlock()
	cfg.productMutex.Lock()
	defer cfg.productMutex.Unlock()
	if cfg.productService == nil {
		if cfg.DB == nil || cfg.DBConn == nil {
			cfg.productService = NewProductService(nil, nil)
		} else {
			cfg.productService = NewProductService(cfg.DB, cfg.DBConn)
		}
	}
	return cfg.productService
}

// handleProductError handles product-specific errors with proper logging and responses.
// It categorizes errors by type and responds with appropriate HTTP status codes and messages.
// All errors are logged with context information for debugging.
func (cfg *HandlersProductConfig) handleProductError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if appErr, ok := err.(*handlers.AppError); ok {
		switch appErr.Code {
		case "transaction_error", "update_failed", "commit_error", "create_product_error", "delete_product_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "product_not_found":
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

// --- Request/Response Structs ---

// ProductRequest represents the data structure for creating or updating a product.
// It includes all product fields with optional ID for updates and optional IsActive for status changes.
type ProductRequest struct {
	ID          string  `json:"id,omitempty"`
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// FilterProductsRequest represents the criteria for filtering products.
// All fields are optional and use nullable types to distinguish between unset and zero values.
type FilterProductsRequest struct {
	CategoryID utils.NullString  `json:"category_id,omitempty"`
	IsActive   utils.NullBool    `json:"is_active,omitempty"`
	MinPrice   utils.NullFloat64 `json:"min_price,omitempty"`
	MaxPrice   utils.NullFloat64 `json:"max_price,omitempty"`
}

// productResponse represents the standard response structure for product operations.
// It includes a message describing the operation result and the product ID when applicable.
type productResponse struct {
	Message   string `json:"message"`
	ProductID string `json:"product_id"`
}
