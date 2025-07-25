// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
)

// order_wrapper.go: Provides order handler configuration, service initialization, error handling, and defines request/response structures for order operations.

// HandlersOrderConfig holds the configuration and dependencies for order handlers.
// Manages the order service lifecycle and provides thread-safe access to the service instance.
type HandlersOrderConfig struct {
	*handlers.Config
	Logger       handlers.HandlerLogger
	orderService OrderService
	orderMutex   sync.RWMutex
}

// InitOrderService initializes the order service with the current configuration.
// Validates that both DB and DBConn are set before creating the service. Returns an error if either dependency is missing.
func (cfg *HandlersOrderConfig) InitOrderService() error {
	if cfg.Config == nil {
		return errors.New("handlers config not initialized")
	}
	if cfg.DB == nil {
		return errors.New("database not initialized")
	}
	if cfg.DBConn == nil {
		return errors.New("database connection not initialized")
	}
	cfg.orderMutex.Lock()
	defer cfg.orderMutex.Unlock()
	cfg.orderService = NewOrderService(cfg.DB, cfg.DBConn)

	// Set Logger if not already set
	if cfg.Logger == nil {
		cfg.Logger = cfg.Config // Config implements HandlerLogger
	}

	return nil
}

// GetOrderService returns the order service instance, initializing it if necessary.
// Uses a double-checked locking pattern for thread-safe lazy initialization. If dependencies are missing, creates a service with nil dependencies.
func (cfg *HandlersOrderConfig) GetOrderService() OrderService {
	cfg.orderMutex.RLock()
	if cfg.orderService != nil {
		defer cfg.orderMutex.RUnlock()
		return cfg.orderService
	}
	cfg.orderMutex.RUnlock()
	cfg.orderMutex.Lock()
	defer cfg.orderMutex.Unlock()
	if cfg.orderService == nil {
		if cfg.Config == nil || cfg.DB == nil || cfg.DBConn == nil {
			cfg.orderService = NewOrderService(nil, nil)
		} else {
			cfg.orderService = NewOrderService(cfg.DB, cfg.DBConn)
		}
	}
	return cfg.orderService
}

// handleOrderError handles order-specific errors with proper logging and responses.
// Categorizes errors by type and responds with appropriate HTTP status codes and messages. All errors are logged with context information for debugging.
func (cfg *HandlersOrderConfig) handleOrderError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	var appErr *handlers.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "transaction_error", "update_failed", "commit_error", "create_order_error", "delete_order_error", "create_order_item_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "order_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "invalid_request", "invalid_status", "quantity_overflow":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "unauthorized":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusForbidden, appErr.Message)
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

// CreateOrderRequest represents the data structure for creating a new order.
type CreateOrderRequest struct {
	Items             []OrderItemInput `json:"items"`
	PaymentMethod     string           `json:"payment_method"`
	ShippingAddress   string           `json:"shipping_address"`
	ContactPhone      string           `json:"contact_phone"`
	ExternalPaymentID string           `json:"external_payment_id,omitempty"`
	TrackingNumber    string           `json:"tracking_number,omitempty"`
}

// OrderItemInput represents an item in an order creation request.
type OrderItemInput struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// UpdateOrderStatusRequest represents the data structure for updating order status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// OrderItemResponse represents an order item in responses.
type OrderItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
}

// UserOrderResponse represents a user's order in list responses.
type UserOrderResponse struct {
	OrderID         string              `json:"order_id"`
	TotalAmount     string              `json:"total_amount"`
	Status          string              `json:"status"`
	PaymentMethod   string              `json:"payment_method,omitempty"`
	TrackingNumber  string              `json:"tracking_number,omitempty"`
	ShippingAddress string              `json:"shipping_address,omitempty"`
	ContactPhone    string              `json:"contact_phone,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	Items           []OrderItemResponse `json:"items"`
}

// OrderResponse represents the standard response structure for order operations.
type OrderResponse struct {
	Message string `json:"message"`
	OrderID string `json:"order_id"`
}

// OrderDetailResponse represents a detailed order response with items.
type OrderDetailResponse struct {
	Order database.Order       `json:"order"`
	Items []database.OrderItem `json:"items"`
}
