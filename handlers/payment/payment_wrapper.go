package paymenthandlers

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/stripe/stripe-go/v82"
)

// HandlersPaymentConfig contains the configuration for payment handlers
// It embeds the base handlers config and provides access to the payment service
// for business logic operations
type HandlersPaymentConfig struct {
	*handlers.HandlersConfig
	Logger         handlers.HandlerLogger
	paymentService PaymentService
	paymentMutex   sync.RWMutex
}

// InitPaymentService initializes the payment service with the current configuration
// This method should be called during application startup
func (cfg *HandlersPaymentConfig) InitPaymentService() error {
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
	if cfg.DBConn == nil {
		return errors.New("database connection not initialized")
	}
	if cfg.StripeSecretKey == "" {
		return errors.New("stripe secret key not configured")
	}

	cfg.paymentMutex.Lock()
	defer cfg.paymentMutex.Unlock()

	cfg.paymentService = NewPaymentService(
		cfg.DB,
		cfg.DBConn,
		cfg.StripeSecretKey,
	)

	// Set Logger if not already set
	if cfg.Logger == nil {
		cfg.Logger = cfg.HandlersConfig // HandlersConfig implements HandlerLogger
	}

	return nil
}

// GetPaymentService returns the payment service instance, initializing it if necessary
// This method is thread-safe and will initialize the service on first access
func (cfg *HandlersPaymentConfig) GetPaymentService() PaymentService {
	cfg.paymentMutex.RLock()
	if cfg.paymentService != nil {
		defer cfg.paymentMutex.RUnlock()
		return cfg.paymentService
	}
	cfg.paymentMutex.RUnlock()

	// Need to initialize, acquire write lock
	cfg.paymentMutex.Lock()
	defer cfg.paymentMutex.Unlock()

	// Double-check pattern in case another goroutine initialized it
	if cfg.paymentService == nil {
		// Validate that the embedded config is not nil before accessing its fields
		if cfg.HandlersConfig == nil || cfg.APIConfig == nil || cfg.DB == nil {
			// Return a default service that will fail gracefully when used
			cfg.paymentService = NewPaymentService(nil, nil, "")
		} else {
			cfg.paymentService = NewPaymentService(
				cfg.DB,
				cfg.DBConn,
				cfg.StripeSecretKey,
			)
		}
	}

	return cfg.paymentService
}

// SetupStripeAPI sets up the Stripe API key (kept for backward compatibility)
func (cfg *HandlersPaymentConfig) SetupStripeAPI() {
	stripe.Key = cfg.StripeSecretKey
}

// handlePaymentError handles payment-specific errors with proper logging and responses
// It categorizes errors and provides appropriate HTTP status codes and messages
func (cfg *HandlersPaymentConfig) handlePaymentError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if appErr, ok := err.(*handlers.AppError); ok {
		switch appErr.Code {
		case "invalid_request", "missing_order_id", "missing_user_id", "invalid_currency", "payment_exists":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, nil)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "order_not_found", "payment_not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "unauthorized":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusForbidden, appErr.Message)
		case "invalid_order_status", "invalid_status", "invalid_amount", "invalid_payment":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		case "database_error", "transaction_error", "commit_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again later")
		case "stripe_error", "webhook_error":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Payment service error")
		default:
			cfg.Logger.LogHandlerError(ctx, operation, "internal_error", appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "unknown_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// Request/Response types (kept for backward compatibility)
type CreatePaymentIntentRequest struct {
	OrderID  string `json:"order_id"`
	Currency string `json:"currency"`
}

type CreatePaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
}

type ConfirmPaymentRequest struct {
	OrderID string `json:"order_id"`
}

type ConfirmPaymentResponse struct {
	Status string `json:"status"`
}

type GetPaymentResponse struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	UserID            string    `json:"user_id"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type PaymentHistoryItem struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	Amount            string    `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
