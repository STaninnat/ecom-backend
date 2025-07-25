// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
)

// cart_wrapper.go: Defines cart business logic interface, handler config, DTOs, error handling, and service initialization.

// CartService defines the business logic interface for cart operations.
type CartService interface {
	AddItemToUserCart(ctx context.Context, userID string, productID string, quantity int) error
	AddItemToGuestCart(ctx context.Context, sessionID string, productID string, quantity int) error
	GetUserCart(ctx context.Context, userID string) (*models.Cart, error)
	GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error)
	UpdateItemQuantity(ctx context.Context, userID string, productID string, quantity int) error
	UpdateGuestItemQuantity(ctx context.Context, sessionID string, productID string, quantity int) error
	RemoveItem(ctx context.Context, userID string, productID string) error
	RemoveGuestItem(ctx context.Context, sessionID string, productID string) error
	DeleteUserCart(ctx context.Context, userID string) error
	DeleteGuestCart(ctx context.Context, sessionID string) error
	CheckoutUserCart(ctx context.Context, userID string) (*CartCheckoutResult, error)
	CheckoutGuestCart(ctx context.Context, sessionID string, userID string) (*CartCheckoutResult, error)
}

// CartCheckoutResult represents the result of a cart checkout operation.
type CartCheckoutResult struct {
	OrderID string `json:"order_id"`
	Message string `json:"message"`
}

// HandlersCartConfig contains configuration and dependencies for cart handlers.
// Embeds Config, provides logger, cartService, and thread safety.
type HandlersCartConfig struct {
	*handlers.Config
	Logger      handlers.HandlerLogger
	CartService CartService
	CartMutex   sync.RWMutex
}

// InitCartService initializes the cart service with the current configuration.
// Sets the CartService and Logger fields, ensuring thread safety with CartMutex.
// Returns an error if the embedded Config is not initialized.
func (cfg *HandlersCartConfig) InitCartService(service CartService) error {
	if cfg.Config == nil {
		return errors.New("handlers config not initialized")
	}
	cfg.CartMutex.Lock()
	defer cfg.CartMutex.Unlock()
	cfg.CartService = service
	if cfg.Logger == nil {
		cfg.Logger = cfg.Config // Config implements HandlerLogger
	}
	return nil
}

// GetCartService returns the cart service instance in a thread-safe manner.
// Acquires a read lock on CartMutex to safely access the CartService field.
func (cfg *HandlersCartConfig) GetCartService() CartService {
	cfg.CartMutex.RLock()
	service := cfg.CartService
	cfg.CartMutex.RUnlock()
	return service
}

// handleCartError maps service errors to HTTP responses and logs them.
// Inspects the error type and code, logs the error, and sends an appropriate HTTP response.
// Used by cart handlers to provide consistent error handling and logging.
func (cfg *HandlersCartConfig) handleCartError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	var appErr *handlers.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message, appErr.Code)
		case "unauthorized":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusForbidden, appErr.Message, appErr.Code)
		case "invalid_request":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message, appErr.Code)
		case "product_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message, appErr.Code)
		case "item_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message, appErr.Code)
		case "cart_empty":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message, appErr.Code)
		case "insufficient_stock":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message, appErr.Code)
		case "add_failed", "get_failed", "update_failed", "remove_failed", "clear_failed", "get_cart_failed", "save_cart_failed":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error", appErr.Code)
		case "invalid_price", "invalid_quantity", "transaction_error", "create_order_failed", "update_stock_failed", "create_order_item_failed", "commit_failed":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error", appErr.Code)
		case "cart_full":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message, appErr.Code)
		case "cart_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message, appErr.Code)
		case "database_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, appErr.Message, appErr.Code)
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error", appErr.Code)
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error", "internal_error")
	}
}

// handleCartItemOperation is a shared helper for add/update item handlers (user/guest)
func (cfg *HandlersCartConfig) handleCartItemOperation(
	w http.ResponseWriter,
	r *http.Request,
	id string,
	parseIDErrMsg string,
	parseReq func(*http.Request) (string, string, int, error),
	serviceCall func(ctx context.Context, id, productID string, quantity int) error,
	opName string,
	successMsg string,
	responseMsg string,
) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	if id == "" {
		cfg.Logger.LogHandlerError(ctx, opName, "missing id", parseIDErrMsg, ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, parseIDErrMsg)
		return
	}

	productID, _, quantity, err := parseReq(r)
	if err != nil {
		cfg.Logger.LogHandlerError(ctx, opName, "invalid request body", "Failed to parse body", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if productID == "" || quantity <= 0 {
		cfg.Logger.LogHandlerError(ctx, opName, "missing fields", "Required fields are missing", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	if err := serviceCall(ctx, id, productID, quantity); err != nil {
		cfg.handleCartError(w, r, err, opName, ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, opName, successMsg, ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{Message: responseMsg})
}

// CartItemRequest is the DTO for cart item operations (add/update).
// Fields:
//   - ProductID: required, identifies the product
//   - Quantity: required, number of items to add/update
//
// Validation is performed in the handler layer.
type CartItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CartUpdateRequest is the DTO for updating cart items in the cart.
// Fields:
//   - ProductID: required, identifies the product
//   - Quantity: required, new quantity for the item
//
// Validation is performed in the handler layer.
type CartUpdateRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CartResponse represents a standard response for cart operations.
// Fields:
//   - Message: operation result message
//   - OrderID: present for checkout responses
//
// Used for API responses to cart actions.
type CartResponse struct {
	Message string `json:"message"`
	OrderID string `json:"order_id,omitempty"`
}

// NewCartServiceWithDeps creates a new cart service with all required dependencies.
// Convenience function for initialization in main or tests.
// Parameters:
//   - cartMongo: MongoDB adapter for cart persistence
//   - db: SQL database queries for product/order
//   - dbConn: SQL database connection
//   - redisClient: Redis client for caching
//
// Returns a CartService implementation.
func NewCartServiceWithDeps(
	cartMongo *intmongo.CartMongo,
	db *database.Queries,
	dbConn *sql.DB,
	redisClient redis.Cmdable,
) CartService {
	return &cartServiceImpl{
		cartMongo: NewCartMongoAdapter(cartMongo),
		product:   NewProductAdapter(db),
		order:     NewOrderAdapter(db),
		dbConn:    NewDBConnAdapter(dbConn),
		redis:     NewCartRedisAPI(redisClient),
	}
}
