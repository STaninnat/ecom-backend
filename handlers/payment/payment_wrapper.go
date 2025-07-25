// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v82"

	"github.com/STaninnat/ecom-backend/handlers"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user"
)

// payment_wrapper.go: Provides payment handler configuration, service initialization, and error handling.

// HandlersPaymentConfig contains the configuration for payment handlers.
// Embeds the base handlers config and provides access to the payment service for business logic operations.
type HandlersPaymentConfig struct {
	*handlers.Config
	Logger         handlers.HandlerLogger
	paymentService PaymentService
	paymentMutex   sync.RWMutex
}

// InitPaymentService initializes the payment service with the current configuration.
// Validates required dependencies and sets up the service. This method should be called during application startup.
func (cfg *HandlersPaymentConfig) InitPaymentService() error {
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
		cfg.Logger = cfg.Config // Config implements HandlerLogger
	}

	return nil
}

// GetPaymentService returns the payment service instance, initializing it if necessary.
// This method is thread-safe and will initialize the service on first access.
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
		if cfg.Config == nil || cfg.APIConfig == nil || cfg.DB == nil {
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

// SetupStripeAPI sets up the Stripe API key (kept for backward compatibility).
func (cfg *HandlersPaymentConfig) SetupStripeAPI() {
	stripe.Key = cfg.StripeSecretKey
}

// handlePaymentError handles payment-specific errors with proper logging and responses.
// Categorizes errors and provides appropriate HTTP status codes and messages. All errors are logged with context information for debugging.
func (cfg *HandlersPaymentConfig) handlePaymentError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	codeMap := map[string]userhandlers.ErrorResponseConfig{
		"invalid_request":      {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
		"missing_order_id":     {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
		"missing_user_id":      {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
		"invalid_currency":     {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
		"payment_exists":       {Status: http.StatusBadRequest, Message: "", UseAppErr: false},
		"order_not_found":      {Status: http.StatusNotFound, Message: "", UseAppErr: true},
		"payment_not_found":    {Status: http.StatusNotFound, Message: "", UseAppErr: true},
		"unauthorized":         {Status: http.StatusForbidden, Message: "", UseAppErr: true},
		"invalid_order_status": {Status: http.StatusBadRequest, Message: "", UseAppErr: true},
		"invalid_status":       {Status: http.StatusBadRequest, Message: "", UseAppErr: true},
		"invalid_amount":       {Status: http.StatusBadRequest, Message: "", UseAppErr: true},
		"invalid_payment":      {Status: http.StatusBadRequest, Message: "", UseAppErr: true},
		"database_error":       {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
		"transaction_error":    {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
		"commit_error":         {Status: http.StatusInternalServerError, Message: "Something went wrong, please try again later", UseAppErr: true},
		"stripe_error":         {Status: http.StatusInternalServerError, Message: "Payment service error", UseAppErr: true},
		"webhook_error":        {Status: http.StatusInternalServerError, Message: "Payment service error", UseAppErr: true},
		"unauthorized_payment": {Status: http.StatusForbidden, Message: "", UseAppErr: true},
		"unauthorized_order":   {Status: http.StatusForbidden, Message: "", UseAppErr: true},
		"unauthorized_user":    {Status: http.StatusForbidden, Message: "", UseAppErr: true},
		"user_not_found":       {Status: http.StatusNotFound, Message: "", UseAppErr: true},
	}
	userhandlers.HandleErrorWithCodeMap(cfg.Logger, w, r, err, operation, ip, userAgent, codeMap, http.StatusInternalServerError, "Internal server error")
}

// CreatePaymentIntentRequest represents the request structure for creating a payment intent.
type CreatePaymentIntentRequest struct {
	OrderID  string `json:"order_id"`
	Currency string `json:"currency"`
}

// CreatePaymentIntentResponse represents the response structure for a created payment intent.
type CreatePaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
}

// ConfirmPaymentRequest represents the request structure for confirming a payment.
type ConfirmPaymentRequest struct {
	OrderID string `json:"order_id"`
}

// ConfirmPaymentResponse represents the response structure for a confirmed payment.
type ConfirmPaymentResponse struct {
	Status string `json:"status"`
}

// GetPaymentResponse represents the response structure for payment details.
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

// PaymentHistoryItem represents a payment item in the payment history.
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
