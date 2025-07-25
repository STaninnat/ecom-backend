// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"errors"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/refund"
	"github.com/stripe/stripe-go/v82/webhook"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// payment_service.go: Implements payment database query adapters, transaction handling, and Stripe client interface.

const (
	orderStatusPending = "pending"
)

// PaymentDBQueries provides methods for payment-related database operations.
type PaymentDBQueries interface {
	WithTx(tx PaymentDBTx) PaymentDBQueries
	GetOrderByID(ctx context.Context, id string) (database.Order, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (database.Payment, error)
	GetPaymentByProviderPaymentID(ctx context.Context, providerPaymentID string) (database.Payment, error)
	GetPaymentsByUserID(ctx context.Context, userID string) ([]database.Payment, error)
	GetAllPayments(ctx context.Context) ([]database.Payment, error)
	GetPaymentsByStatus(ctx context.Context, status string) ([]database.Payment, error)
	CreatePayment(ctx context.Context, params database.CreatePaymentParams) error
	UpdatePaymentStatus(ctx context.Context, params database.UpdatePaymentStatusParams) error
	UpdatePaymentStatusByID(ctx context.Context, params database.UpdatePaymentStatusByIDParams) error
	UpdatePaymentStatusByProviderPaymentID(ctx context.Context, params database.UpdatePaymentStatusByProviderPaymentIDParams) error
	UpdateOrderStatus(ctx context.Context, params database.UpdateOrderStatusParams) error
}

// PaymentDBConn defines the interface for beginning database transactions for payment operations.
type PaymentDBConn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (PaymentDBTx, error)
}

// PaymentDBTx defines the interface for a database transaction used in payment operations.
type PaymentDBTx interface {
	Commit() error
	Rollback() error
}

// PaymentDBQueriesAdapter adapts sqlc-generated Queries to the PaymentDBQueries interface.
type PaymentDBQueriesAdapter struct {
	*database.Queries
}

// WithTx returns a new PaymentDBQueries that uses the provided transaction.
func (a *PaymentDBQueriesAdapter) WithTx(tx PaymentDBTx) PaymentDBQueries {
	if tx == nil {
		return nil
	}

	return &PaymentDBQueriesAdapter{a.Queries.WithTx(tx.(*sql.Tx))}
}

// GetOrderByID retrieves an order by its ID.
func (a *PaymentDBQueriesAdapter) GetOrderByID(ctx context.Context, id string) (database.Order, error) {
	return a.Queries.GetOrderByID(ctx, id)
}

// GetPaymentByOrderID retrieves a payment by its order ID.
func (a *PaymentDBQueriesAdapter) GetPaymentByOrderID(ctx context.Context, orderID string) (database.Payment, error) {
	return a.Queries.GetPaymentByOrderID(ctx, orderID)
}

// GetPaymentByProviderPaymentID is a placeholder and not implemented in the real adapter.
func (a *PaymentDBQueriesAdapter) GetPaymentByProviderPaymentID(_ context.Context, _ string) (database.Payment, error) {
	return database.Payment{}, errors.New("GetPaymentByProviderPaymentID not implemented in real adapter")
}

// GetPaymentsByUserID retrieves all payments for a specific user.
func (a *PaymentDBQueriesAdapter) GetPaymentsByUserID(ctx context.Context, userID string) ([]database.Payment, error) {
	return a.Queries.GetPaymentsByUserID(ctx, userID)
}

// GetAllPayments retrieves all payments from the database.
func (a *PaymentDBQueriesAdapter) GetAllPayments(ctx context.Context) ([]database.Payment, error) {
	return a.Queries.GetAllPayments(ctx)
}

// GetPaymentsByStatus retrieves all payments with a specific status.
func (a *PaymentDBQueriesAdapter) GetPaymentsByStatus(ctx context.Context, status string) ([]database.Payment, error) {
	return a.Queries.GetPaymentsByStatus(ctx, status)
}

// CreatePayment creates a new payment record in the database.
func (a *PaymentDBQueriesAdapter) CreatePayment(ctx context.Context, params database.CreatePaymentParams) error {
	_, err := a.Queries.CreatePayment(ctx, params)
	return err
}

// UpdatePaymentStatus updates the status of a payment by its ID.
func (a *PaymentDBQueriesAdapter) UpdatePaymentStatus(ctx context.Context, params database.UpdatePaymentStatusParams) error {
	return a.Queries.UpdatePaymentStatus(ctx, params)
}

// UpdatePaymentStatusByID updates the status of a payment by its ID.
func (a *PaymentDBQueriesAdapter) UpdatePaymentStatusByID(ctx context.Context, params database.UpdatePaymentStatusByIDParams) error {
	return a.Queries.UpdatePaymentStatusByID(ctx, params)
}

// UpdatePaymentStatusByProviderPaymentID updates the status of a payment by its provider payment ID.
func (a *PaymentDBQueriesAdapter) UpdatePaymentStatusByProviderPaymentID(ctx context.Context, params database.UpdatePaymentStatusByProviderPaymentIDParams) error {
	return a.Queries.UpdatePaymentStatusByProviderPaymentID(ctx, params)
}

// UpdateOrderStatus updates the status of an order.
func (a *PaymentDBQueriesAdapter) UpdateOrderStatus(ctx context.Context, params database.UpdateOrderStatusParams) error {
	return a.Queries.UpdateOrderStatus(ctx, params)
}

// PaymentDBConnAdapter adapts a sql.DB to the PaymentDBConn interface.
type PaymentDBConnAdapter struct {
	*sql.DB
}

// BeginTx begins a new database transaction.
func (a *PaymentDBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (PaymentDBTx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	return tx, err
}

// StripeClient abstracts Stripe operations for testability.
// (If you use mockery, otherwise define a manual mock in tests)
//
//go:generate mockery --name=StripeClient --output=./mocks --case=underscore
type StripeClient interface {
	CreatePaymentIntent(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
	GetPaymentIntent(id string) (*stripe.PaymentIntent, error)
	CreateRefund(params *stripe.RefundParams) (*stripe.Refund, error)
	ParseWebhook(payload []byte, sigHeader, secret string) (stripe.Event, error)
}

// realStripeClient implements StripeClient using the stripe-go SDK (used in production).
type realStripeClient struct{}

func (c *realStripeClient) CreatePaymentIntent(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	return paymentintent.New(params)
}
func (c *realStripeClient) GetPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(id, nil)
}
func (c *realStripeClient) CreateRefund(params *stripe.RefundParams) (*stripe.Refund, error) {
	return refund.New(params)
}
func (c *realStripeClient) ParseWebhook(payload []byte, sigHeader, secret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, secret)
}

// --- Service Implementation ---
type paymentServiceImpl struct {
	db     PaymentDBQueries
	dbConn PaymentDBConn
	apiKey string
	stripe StripeClient
}

// PaymentService defines the business logic interface for payment operations.
// Provides methods for creating, confirming, retrieving, and refunding payments, as well as handling webhooks.
type PaymentService interface {
	CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error)
	ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error)
	GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error)
	GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error)
	GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error)
	RefundPayment(ctx context.Context, params RefundPaymentParams) error
	HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error
}

// CreatePaymentParams contains parameters for creating a payment.
type CreatePaymentParams struct {
	OrderID  string
	UserID   string
	Currency string
}

// CreatePaymentResult represents the result of creating a payment.
type CreatePaymentResult struct {
	PaymentID    string
	ClientSecret string
}

// ConfirmPaymentParams represents the parameters for confirming a payment.
type ConfirmPaymentParams struct {
	OrderID string
	UserID  string
}

// ConfirmPaymentResult represents the result of confirming a payment.
type ConfirmPaymentResult struct {
	Status string
}

// GetPaymentResult represents the result of retrieving a payment.
type GetPaymentResult struct {
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

// RefundPaymentParams represents the parameters for refunding a payment.
type RefundPaymentParams struct {
	OrderID string
	UserID  string
}

// NewPaymentService creates a new PaymentService with the provided database query and connection adapters.
// Returns a PaymentService implementation.
func NewPaymentService(db *database.Queries, dbConn *sql.DB, apiKey string) PaymentService {
	return &paymentServiceImpl{
		db:     &PaymentDBQueriesAdapter{db},
		dbConn: &PaymentDBConnAdapter{dbConn},
		apiKey: apiKey,
		stripe: &realStripeClient{}, // use real client by default
	}
}

// CreatePayment creates a new payment intent and records it in the database.
// Validates the request, creates a Stripe payment intent, and records the payment in a transaction.
func (s *paymentServiceImpl) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	if params.OrderID == "" || params.UserID == "" || params.Currency == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Missing required fields"}
	}

	// Validate currency
	if !s.isValidCurrency(params.Currency) {
		return nil, &handlers.AppError{Code: "invalid_currency", Message: "Unsupported currency"}
	}

	// Get order and validate ownership
	order, err := s.db.GetOrderByID(ctx, params.OrderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "order_not_found", Message: "Order not found", Err: err}
	}
	if order.UserID != params.UserID {
		return nil, &handlers.AppError{Code: "unauthorized", Message: "Order does not belong to user"}
	}
	if order.Status != orderStatusPending {
		return nil, &handlers.AppError{Code: "invalid_order_status", Message: "Order already paid or invalid"}
	}

	// Check if payment already exists for this order
	existingPayment, err := s.db.GetPaymentByOrderID(ctx, params.OrderID)
	if err == nil && existingPayment.ID != "" {
		return nil, &handlers.AppError{Code: "payment_exists", Message: "Payment already exists for this order"}
	}

	// Parse amount with proper precision
	totalAmount, err := strconv.ParseFloat(order.TotalAmount, 64)
	if err != nil {
		return nil, &handlers.AppError{Code: "invalid_amount", Message: "Invalid total amount", Err: err}
	}

	// Convert to smallest currency unit (cents for USD, etc.) with proper rounding
	amountInSmallestUnit := int64(math.Round(totalAmount * 100))

	// Create Stripe payment intent
	stripe.Key = s.apiKey
	metadata := map[string]string{
		"order_id": order.ID,
		"user_id":  params.UserID,
	}
	stripeParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInSmallestUnit),
		Currency: stripe.String(params.Currency),
		Metadata: metadata,
	}
	intent, err := s.stripe.CreatePaymentIntent(stripeParams)
	if err != nil {
		return nil, &handlers.AppError{Code: "stripe_error", Message: "Failed to create payment intent", Err: err}
	}

	// Record payment in database
	paymentID := utils.NewUUIDString()
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)

	err = queries.CreatePayment(ctx, database.CreatePaymentParams{
		ID:                paymentID,
		OrderID:           order.ID,
		UserID:            params.UserID,
		Amount:            order.TotalAmount,
		Currency:          params.Currency,
		Status:            "created",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString(intent.ID),
		CreatedAt:         timeNow,
		UpdatedAt:         timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to record payment", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return &CreatePaymentResult{
		PaymentID:    paymentID,
		ClientSecret: intent.ClientSecret,
	}, nil
}

// isValidCurrency checks if the currency is supported.
func (s *paymentServiceImpl) isValidCurrency(currency string) bool {
	valid := map[string]struct{}{
		"USD": {}, "EUR": {}, "GBP": {}, "JPY": {}, "CAD": {}, "AUD": {}, "CHF": {}, "CNY": {}, "SEK": {}, "NZD": {},
	}
	_, ok := valid[currency]
	return ok
}

// ConfirmPayment confirms a payment and updates its status.
// Validates the payment, checks Stripe status, and updates both payment and order status in a transaction.
func (s *paymentServiceImpl) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	if params.OrderID == "" || params.UserID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Missing required fields"}
	}

	// Get payment and validate ownership
	payment, err := s.db.GetPaymentByOrderID(ctx, params.OrderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "payment_not_found", Message: "Payment not found", Err: err}
	}
	if payment.UserID != params.UserID {
		return nil, &handlers.AppError{Code: "unauthorized", Message: "Payment does not belong to user"}
	}

	// Get payment intent from Stripe
	if !payment.ProviderPaymentID.Valid {
		return nil, &handlers.AppError{Code: "invalid_payment", Message: "Missing provider payment ID"}
	}

	stripe.Key = s.apiKey
	pi, err := s.stripe.GetPaymentIntent(payment.ProviderPaymentID.String)
	if err != nil {
		return nil, &handlers.AppError{Code: "stripe_error", Message: "Failed to fetch payment intent", Err: err}
	}

	// Check for refunds first
	hasRefund := false
	// No-op: refund status is handled by webhook events

	// Determine new status
	var newStatus string
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		if hasRefund {
			newStatus = "refunded"
		} else {
			newStatus = "succeeded"
		}
	case stripe.PaymentIntentStatusCanceled:
		newStatus = "cancelled"
	case stripe.PaymentIntentStatusRequiresPaymentMethod, stripe.PaymentIntentStatusRequiresConfirmation:
		newStatus = orderStatusPending
	case stripe.PaymentIntentStatusRequiresAction:
		newStatus = orderStatusPending
	case stripe.PaymentIntentStatusRequiresCapture:
		newStatus = orderStatusPending
	case stripe.PaymentIntentStatusProcessing:
		newStatus = orderStatusPending
	default:
		newStatus = "failed"
	}

	// Update payment and order status
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)

	err = queries.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
		ID:        payment.ID,
		Status:    newStatus,
		UpdatedAt: timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to update payment status", Err: err}
	}

	// Update order status if payment succeeded
	statusMap := map[string]string{
		"succeeded": "paid",
		"refunded":  "refunded",
	}
	if updatedStatus, ok := statusMap[newStatus]; ok {
		err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
			ID:        payment.OrderID,
			Status:    updatedStatus,
			UpdatedAt: timeNow,
		})
		if err != nil {
			return nil, &handlers.AppError{Code: "database_error", Message: "Failed to update order status", Err: err}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return &ConfirmPaymentResult{Status: newStatus}, nil
}

// GetPayment retrieves a payment by order ID.
// Validates ownership and returns payment details or an error.
func (s *paymentServiceImpl) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	if orderID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Order ID is required"}
	}

	payment, err := s.db.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "payment_not_found", Message: "Payment not found", Err: err}
	}

	if payment.UserID != userID {
		return nil, &handlers.AppError{Code: "unauthorized", Message: "Payment does not belong to user"}
	}

	amount, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		return nil, &handlers.AppError{Code: "invalid_amount", Message: "Invalid payment amount", Err: err}
	}

	return &GetPaymentResult{
		ID:                payment.ID,
		OrderID:           payment.OrderID,
		UserID:            payment.UserID,
		Amount:            amount,
		Currency:          payment.Currency,
		Status:            payment.Status,
		Provider:          payment.Provider,
		ProviderPaymentID: payment.ProviderPaymentID.String,
		CreatedAt:         payment.CreatedAt,
	}, nil
}

// GetPaymentHistory retrieves payment history for a user.
// Returns a list of payment history items or an error.
func (s *paymentServiceImpl) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	if userID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}

	payments, err := s.db.GetPaymentsByUserID(ctx, userID)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to fetch payment history", Err: err}
	}

	result := make([]PaymentHistoryItem, 0, len(payments))
	for _, p := range payments {
		result = append(result, PaymentHistoryItem{
			ID:                p.ID,
			OrderID:           p.OrderID,
			Amount:            p.Amount,
			Currency:          p.Currency,
			Status:            p.Status,
			Provider:          p.Provider,
			ProviderPaymentID: p.ProviderPaymentID.String,
			CreatedAt:         p.CreatedAt,
		})
	}

	return result, nil
}

// GetAllPayments retrieves all payments with optional status filter.
// Returns a list of payment history items or an error.
func (s *paymentServiceImpl) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	var payments []database.Payment
	var err error

	if status == "all" {
		payments, err = s.db.GetAllPayments(ctx)
	} else {
		payments, err = s.db.GetPaymentsByStatus(ctx, status)
	}

	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to fetch payments", Err: err}
	}

	result := make([]PaymentHistoryItem, 0, len(payments))
	for _, p := range payments {
		result = append(result, PaymentHistoryItem{
			ID:                p.ID,
			OrderID:           p.OrderID,
			Amount:            p.Amount,
			Currency:          p.Currency,
			Status:            p.Status,
			Provider:          p.Provider,
			ProviderPaymentID: p.ProviderPaymentID.String,
			CreatedAt:         p.CreatedAt,
		})
	}

	return result, nil
}

// RefundPayment processes a refund for a payment.
// Validates the payment, processes the refund with Stripe, and updates both payment and order status in a transaction.
func (s *paymentServiceImpl) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	if params.OrderID == "" || params.UserID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Missing required fields"}
	}

	// Get payment and validate ownership
	payment, err := s.db.GetPaymentByOrderID(ctx, params.OrderID)
	if err != nil {
		return &handlers.AppError{Code: "payment_not_found", Message: "Payment not found", Err: err}
	}
	if payment.UserID != params.UserID {
		return &handlers.AppError{Code: "unauthorized", Message: "Payment does not belong to user"}
	}

	if payment.Status != "succeeded" {
		return &handlers.AppError{Code: "invalid_status", Message: "Payment cannot be refunded"}
	}

	if !payment.ProviderPaymentID.Valid {
		return &handlers.AppError{Code: "invalid_payment", Message: "Missing provider payment ID"}
	}

	// Process refund with Stripe
	refundParams := &stripe.RefundParams{
		PaymentIntent: stripe.String(payment.ProviderPaymentID.String),
	}
	_, err = s.stripe.CreateRefund(refundParams)
	if err != nil {
		return &handlers.AppError{Code: "stripe_error", Message: "Failed to process refund", Err: err}
	}

	// Update payment and order status
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)

	err = queries.UpdatePaymentStatusByID(ctx, database.UpdatePaymentStatusByIDParams{
		ID:     payment.ID,
		Status: "refunded",
	})
	if err != nil {
		return &handlers.AppError{Code: "database_error", Message: "Failed to update payment status", Err: err}
	}

	err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
		ID:        payment.OrderID,
		Status:    "cancelled",
		UpdatedAt: timeNow,
	})
	if err != nil {
		return &handlers.AppError{Code: "database_error", Message: "Failed to update order status", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// PaymentError represents payment-specific errors.
type PaymentError = handlers.AppError

// HandleWebhook processes Stripe webhook events.
// Validates the webhook signature, processes different event types, and updates payment statuses in a transaction.
func (s *paymentServiceImpl) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	stripe.Key = s.apiKey
	event, err := s.stripe.ParseWebhook(payload, signature, secret)
	if err != nil {
		return &handlers.AppError{Code: "webhook_error", Message: "Signature verification failed", Err: err}
	}

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = tx.Rollback()
	}()

	queries := s.db.WithTx(tx)

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return &handlers.AppError{Code: "webhook_error", Message: "Bad payment intent", Err: err}
		}
		if _, err := s.db.GetPaymentByProviderPaymentID(ctx, pi.ID); err != nil {
			return &handlers.AppError{Code: "payment_not_found", Message: "Payment not found", Err: err}
		}
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "succeeded",
			ProviderPaymentID: utils.ToNullString(pi.ID),
		})
		if err != nil {
			return &handlers.AppError{Code: "database_error", Message: "Failed to update payment", Err: err}
		}

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return &handlers.AppError{Code: "webhook_error", Message: "Bad payment intent", Err: err}
		}
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "failed",
			ProviderPaymentID: utils.ToNullString(pi.ID),
		})
		if err != nil {
			return &handlers.AppError{Code: "database_error", Message: "Failed to update payment", Err: err}
		}

	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return &handlers.AppError{Code: "webhook_error", Message: "Bad payment intent", Err: err}
		}
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "cancelled",
			ProviderPaymentID: utils.ToNullString(pi.ID),
		})
		if err != nil {
			return &handlers.AppError{Code: "database_error", Message: "Failed to update payment", Err: err}
		}

	case "charge.refunded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			return &handlers.AppError{Code: "webhook_error", Message: "Bad charge", Err: err}
		}
		// Find payment by charge ID and update status
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "refunded",
			ProviderPaymentID: utils.ToNullString(charge.PaymentIntent.ID),
		})
		if err != nil {
			return &handlers.AppError{Code: "database_error", Message: "Failed to update payment", Err: err}
		}

	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}
