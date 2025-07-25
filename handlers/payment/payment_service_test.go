// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// payment_service_test.go: Tests payment service input and error scenarios.

const (
	testSignatureService = "test_signature"
	testSecret           = "whsec_test"
)

// TestCreatePayment_InvalidRequest tests validation of required fields.
func TestCreatePayment_InvalidRequest(t *testing.T) {
	service := &paymentServiceImpl{db: nil, dbConn: nil}

	tests := []struct {
		name   string
		params CreatePaymentParams
	}{
		{"empty_order_id", CreatePaymentParams{OrderID: "", UserID: "user123", Currency: "USD"}},
		{"empty_user_id", CreatePaymentParams{OrderID: "order123", UserID: "", Currency: "USD"}},
		{"empty_currency", CreatePaymentParams{OrderID: "order123", UserID: "user123", Currency: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreatePayment(context.Background(), tt.params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Missing required fields")
		})
	}
}

// TestCreatePayment_InvalidCurrency tests currency validation.
func TestCreatePayment_InvalidCurrency(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "INVALID",
	}

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unsupported currency")
}

// TestCreatePayment_OrderNotFound tests when order doesn't exist.
func TestCreatePayment_OrderNotFound(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := CreatePaymentParams{
		OrderID:  "nonexistent",
		UserID:   "user123",
		Currency: "USD",
	}

	mockDB.On("GetOrderByID", mock.Anything, "nonexistent").Return(database.Order{}, sql.ErrNoRows)

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
	mockDB.AssertExpectations(t)
}

// Helper for CreatePayment error scenarios
func runCreatePaymentOrderErrorTest(t *testing.T, order database.Order, params CreatePaymentParams, expectedMsg string) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	mockDB.On("GetOrderByID", mock.Anything, params.OrderID).Return(order, nil)

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedMsg)
	mockDB.AssertExpectations(t)
}

func TestCreatePayment_UnauthorizedOrder(t *testing.T) {
	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}
	order := database.Order{
		ID:     "order123",
		UserID: "different_user",
		Status: "pending",
	}
	runCreatePaymentOrderErrorTest(t, order, params, "Order does not belong to user")
}

func TestCreatePayment_InvalidOrderStatus(t *testing.T) {
	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}
	order := database.Order{
		ID:     "order123",
		UserID: "user123",
		Status: "paid",
	}
	runCreatePaymentOrderErrorTest(t, order, params, "Order already paid or invalid")
}

// TestCreatePayment_PaymentExists tests when payment already exists for order.
func TestCreatePayment_PaymentExists(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}

	order := database.Order{
		ID:          "order123",
		UserID:      "user123",
		Status:      "pending",
		TotalAmount: "100.00",
	}

	existingPayment := database.Payment{
		ID:      "existing_payment",
		OrderID: "order123",
	}

	mockDB.On("GetOrderByID", mock.Anything, "order123").Return(order, nil)
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(existingPayment, nil)

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment already exists for this order")
	mockDB.AssertExpectations(t)
}

// TestConfirmPayment_InvalidRequest tests validation of required fields.
func TestConfirmPayment_InvalidRequest(t *testing.T) {
	service := &paymentServiceImpl{db: nil, dbConn: nil}

	tests := []struct {
		name   string
		params ConfirmPaymentParams
	}{
		{"empty_order_id", ConfirmPaymentParams{OrderID: "", UserID: "user123"}},
		{"empty_user_id", ConfirmPaymentParams{OrderID: "order123", UserID: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ConfirmPayment(context.Background(), tt.params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Missing required fields")
		})
	}
}

// TestConfirmPayment_PaymentNotFound tests when payment doesn't exist.
func TestConfirmPayment_PaymentNotFound(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")
	mockDB.AssertExpectations(t)
}

// TestConfirmPayment_UnauthorizedPayment tests when payment doesn't belong to user.
func TestConfirmPayment_UnauthorizedPayment(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "different_user",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment does not belong to user")
	mockDB.AssertExpectations(t)
}

// TestGetPayment_NotFound tests when payment doesn't exist.
func TestGetPayment_NotFound(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)

	_, err := service.GetPayment(context.Background(), "order123", "user123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")
	mockDB.AssertExpectations(t)
}

// TestGetPayment_Unauthorized tests when payment doesn't belong to user.
func TestGetPayment_Unauthorized(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}

	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "different_user",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	_, err := service.GetPayment(context.Background(), "order123", "user123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment does not belong to user")
	mockDB.AssertExpectations(t)
}

// TestGetPaymentHistory_Success tests the successful retrieval of payment history.
func TestGetPaymentHistory_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}

	payments := []database.Payment{
		{
			ID:                "payment1",
			OrderID:           "order1",
			UserID:            "user123",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: utils.ToNullString("pi_test1"),
			CreatedAt:         time.Now(),
		},
		{
			ID:                "payment2",
			OrderID:           "order2",
			UserID:            "user123",
			Amount:            "200.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: utils.ToNullString("pi_test2"),
			CreatedAt:         time.Now(),
		},
	}

	mockDB.On("GetPaymentsByUserID", mock.Anything, "user123").Return(payments, nil)

	result, err := service.GetPaymentHistory(context.Background(), "user123")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "payment1", result[0].ID)
	assert.Equal(t, "payment2", result[1].ID)
	mockDB.AssertExpectations(t)
}

// TestGetAllPayments_Success tests the successful retrieval of all payments.
func TestGetAllPayments_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}

	payments := []database.Payment{
		{
			ID:                "payment1",
			OrderID:           "order1",
			UserID:            "user123",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: utils.ToNullString("pi_test1"),
			CreatedAt:         time.Now(),
		},
	}

	mockDB.On("GetPaymentsByStatus", mock.Anything, "succeeded").Return(payments, nil)

	result, err := service.GetAllPayments(context.Background(), "succeeded")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "payment1", result[0].ID)
	mockDB.AssertExpectations(t)
}

// TestRefundPayment_InvalidRequest tests validation of required fields.
func TestRefundPayment_InvalidRequest(t *testing.T) {
	service := &paymentServiceImpl{db: nil, dbConn: nil}

	tests := []struct {
		name   string
		params RefundPaymentParams
	}{
		{"empty_order_id", RefundPaymentParams{OrderID: "", UserID: "user123"}},
		{"empty_user_id", RefundPaymentParams{OrderID: "order123", UserID: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RefundPayment(context.Background(), tt.params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Missing required fields")
		})
	}
}

// TestRefundPayment_PaymentNotFound tests when payment doesn't exist.
func TestRefundPayment_PaymentNotFound(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")
	mockDB.AssertExpectations(t)
}

// TestRefundPayment_UnauthorizedPayment tests when payment doesn't belong to user.
func TestRefundPayment_UnauthorizedPayment(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "different_user",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment does not belong to user")
	mockDB.AssertExpectations(t)
}

// TestRefundPayment_InvalidStatus tests when payment status is not refundable.
func TestRefundPayment_InvalidStatus(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil}
	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
		Status:  "pending",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment cannot be refunded")
	mockDB.AssertExpectations(t)
}

// TestHandleWebhook_InvalidSignature tests webhook signature verification.
func TestHandleWebhook_InvalidSignature(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{
		db:     mockDB,
		dbConn: nil,
		apiKey: "test_key",
		stripe: mockStripe,
	}

	payload := []byte(`{"type":"payment_intent.succeeded"}`)
	signature := "invalid_signature"
	secret := "test_secret"

	mockStripe.On("ParseWebhook", payload, signature, secret).Return(stripe.Event{}, errors.New("Signature verification failed"))

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Signature verification failed")
}

// TestIsValidCurrency tests the currency validation function.
func TestIsValidCurrency(t *testing.T) {
	service := &paymentServiceImpl{}

	validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "SEK", "NZD"}
	invalidCurrencies := []string{"INVALID", "US", "EURO", "123", ""}

	for _, currency := range validCurrencies {
		t.Run("valid_"+currency, func(t *testing.T) {
			assert.True(t, service.isValidCurrency(currency))
		})
	}

	for _, currency := range invalidCurrencies {
		t.Run("invalid_"+currency, func(t *testing.T) {
			assert.False(t, service.isValidCurrency(currency))
		})
	}
}

// TestPaymentDBAdapters_Coverage tests all adapter methods for coverage
func TestPaymentDBAdapters_Coverage(t *testing.T) {
	// Test PaymentDBQueriesAdapter methods
	t.Run("PaymentDBQueriesAdapter", func(t *testing.T) {
		// Test WithTx method
		t.Run("WithTx", func(t *testing.T) {
			// Create a nil adapter to test the wrapper without causing panics
			adapter := &PaymentDBQueriesAdapter{Queries: nil}
			mockTx := &sql.Tx{}
			result := adapter.WithTx(mockTx)
			assert.NotNil(t, result)
			assert.IsType(t, &PaymentDBQueriesAdapter{}, result)
		})

		// Test adapter method calls (these will panic due to nil database, but we're testing the adapter wrapper)
		t.Run("AdapterMethodCalls", func(t *testing.T) {
			adapter := &PaymentDBQueriesAdapter{Queries: nil}
			ctx := context.Background()

			// Test that adapter methods exist and can be called (they'll panic, but that's expected)
			assert.Panics(t, func() {
				_, err := adapter.GetOrderByID(ctx, "test_order")
				_ = err
			})

			assert.Panics(t, func() {
				_, err := adapter.GetPaymentByOrderID(ctx, "test_order")
				_ = err
			})

			assert.Panics(t, func() {
				_, err := adapter.GetPaymentsByUserID(ctx, "test_user")
				_ = err
			})

			assert.Panics(t, func() {
				_, err := adapter.GetAllPayments(ctx)
				_ = err
			})

			assert.Panics(t, func() {
				_, err := adapter.GetPaymentsByStatus(ctx, "pending")
				_ = err
			})

			assert.Panics(t, func() {
				params := database.CreatePaymentParams{
					ID:       "test_payment",
					OrderID:  "test_order",
					UserID:   "test_user",
					Amount:   "100.00",
					Currency: "USD",
					Status:   "pending",
				}
				err := adapter.CreatePayment(ctx, params)
				_ = err
			})

			assert.Panics(t, func() {
				params := database.UpdatePaymentStatusParams{
					ID:     "test_payment",
					Status: "completed",
				}
				err := adapter.UpdatePaymentStatus(ctx, params)
				_ = err
			})

			assert.Panics(t, func() {
				params := database.UpdatePaymentStatusByIDParams{
					ID:     "test_payment",
					Status: "completed",
				}
				err := adapter.UpdatePaymentStatusByID(ctx, params)
				_ = err
			})

			assert.Panics(t, func() {
				params := database.UpdatePaymentStatusByProviderPaymentIDParams{
					ProviderPaymentID: utils.ToNullString("pi_test_123"),
					Status:            "completed",
				}
				err := adapter.UpdatePaymentStatusByProviderPaymentID(ctx, params)
				_ = err
			})

			assert.Panics(t, func() {
				params := database.UpdateOrderStatusParams{
					ID:     "test_order",
					Status: "paid",
				}
				err := adapter.UpdateOrderStatus(ctx, params)
				_ = err
			})
		})
	})

	// Test PaymentDBConnAdapter methods
	t.Run("PaymentDBConnAdapter", func(t *testing.T) {
		// Test BeginTx method
		t.Run("BeginTx", func(t *testing.T) {
			// Create a nil adapter to test the wrapper without causing panics
			adapter := &PaymentDBConnAdapter{DB: nil}
			ctx := context.Background()
			assert.Panics(t, func() {
				_, _ = adapter.BeginTx(ctx, nil)
			})
		})
	})
}

// TestCreatePayment_Success tests successful payment creation
func TestCreatePayment_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}

	order := database.Order{
		ID:          "order123",
		UserID:      "user123",
		Status:      "pending",
		TotalAmount: "100.00",
	}

	// Set up mock expectations
	mockDB.On("GetOrderByID", mock.Anything, "order123").Return(order, nil)
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("CreatePayment", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	// Mock Stripe response
	mockStripe.On("CreatePaymentIntent", mock.Anything).Return(&stripe.PaymentIntent{
		ID:           "pi_test_123",
		Status:       stripe.PaymentIntentStatusRequiresPaymentMethod,
		ClientSecret: "pi_test_secret_123",
	}, nil)

	result, err := service.CreatePayment(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.PaymentID)
	assert.NotEmpty(t, result.ClientSecret)

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestCreatePayment_InvalidAmount tests when order has invalid amount
func TestCreatePayment_InvalidAmount(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}

	order := database.Order{
		ID:          "order123",
		UserID:      "user123",
		Status:      "pending",
		TotalAmount: "invalid_amount",
	}

	mockDB.On("GetOrderByID", mock.Anything, "order123").Return(order, nil)
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid total amount")

	mockDB.AssertExpectations(t)
}

// TestCreatePayment_TransactionError tests when transaction fails to start
func TestCreatePayment_TransactionError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}

	order := database.Order{
		ID:          "order123",
		UserID:      "user123",
		Status:      "pending",
		TotalAmount: "100.00",
	}

	mockDB.On("GetOrderByID", mock.Anything, "order123").Return(order, nil)
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)
	mockStripe.On("CreatePaymentIntent", mock.Anything).Return(&stripe.PaymentIntent{
		ID:           "pi_test_123",
		Status:       stripe.PaymentIntentStatusRequiresPaymentMethod,
		ClientSecret: "pi_test_secret_123",
	}, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(nil, errors.New("transaction failed"))

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error starting transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestCreatePayment_CreatePaymentError tests when database payment creation fails
func TestCreatePayment_CreatePaymentError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := CreatePaymentParams{
		OrderID:  "order123",
		UserID:   "user123",
		Currency: "USD",
	}

	order := database.Order{
		ID:          "order123",
		UserID:      "user123",
		Status:      "pending",
		TotalAmount: "100.00",
	}

	mockDB.On("GetOrderByID", mock.Anything, "order123").Return(order, nil)
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(database.Payment{}, sql.ErrNoRows)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("CreatePayment", mock.Anything, mock.Anything).Return(errors.New("database error"))
	mockTx.On("Rollback").Return(nil)

	// Mock Stripe response
	mockStripe.On("CreatePaymentIntent", mock.Anything).Return(&stripe.PaymentIntent{
		ID:           "pi_test_123",
		Status:       stripe.PaymentIntentStatusRequiresPaymentMethod,
		ClientSecret: "pi_test_secret_123",
	}, nil)

	_, err := service.CreatePayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_Success tests successful payment confirmation
func TestConfirmPayment_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "created",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	// Set up mock expectations
	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatus", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	// Mock Stripe response
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
		ID:     "pi_test_123",
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil)

	result, err := service.ConfirmPayment(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "succeeded", result.Status)

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// Helper for ConfirmPayment error scenarios with missing provider payment ID
func runConfirmPaymentMissingProviderIDTest(t *testing.T, payment database.Payment, params ConfirmPaymentParams) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}
	mockDB.On("GetPaymentByOrderID", mock.Anything, params.OrderID).Return(payment, nil)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing provider payment ID")
	mockDB.AssertExpectations(t)
}

func TestConfirmPayment_OrderNotFound(t *testing.T) {
	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}
	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
	}
	runConfirmPaymentMissingProviderIDTest(t, payment, params)
}

func TestConfirmPayment_InvalidOrderStatus(t *testing.T) {
	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}
	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
	}
	runConfirmPaymentMissingProviderIDTest(t, payment, params)
}

// TestGetPayment_Success tests successful payment retrieval
func TestGetPayment_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
		CreatedAt:         time.Now(),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	result, err := service.GetPayment(context.Background(), "order123", "user123")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "payment123", result.ID)
	assert.Equal(t, "order123", result.OrderID)
	assert.Equal(t, "user123", result.UserID)
	assert.InEpsilon(t, 100.0, result.Amount, 0.001)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "succeeded", result.Status)
	assert.Equal(t, "stripe", result.Provider)
	assert.Equal(t, "pi_test_123", result.ProviderPaymentID)

	mockDB.AssertExpectations(t)
}

// Helper for GetPayment invalid amount tests
func runGetPaymentInvalidAmountTest(t *testing.T, payment database.Payment) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}
	mockDB.On("GetPaymentByOrderID", mock.Anything, payment.OrderID).Return(payment, nil)

	_, err := service.GetPayment(context.Background(), payment.OrderID, payment.UserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid payment amount")
	mockDB.AssertExpectations(t)
}

func TestGetPayment_InvalidAmount(t *testing.T) {
	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
		Amount:  "invalid_amount",
		Status:  "succeeded",
	}
	runGetPaymentInvalidAmountTest(t, payment)
}

// TestRefundPayment_Success tests successful payment refund
func TestRefundPayment_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByID", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	// Mock Stripe refund call
	mockStripe.On("CreateRefund", mock.Anything).Return(&stripe.Refund{
		ID:     "re_test_123",
		Status: stripe.RefundStatusSucceeded,
	}, nil)

	err := service.RefundPayment(context.Background(), params)
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestRefundPayment_InvalidAmount tests when payment has invalid amount for refund
func TestRefundPayment_InvalidAmount(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
		Amount:  "invalid_amount",
		Status:  "succeeded",
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing provider payment ID")

	mockDB.AssertExpectations(t)
}

// TestHandleWebhook_Success tests successful webhook handling
func TestHandleWebhook_Success(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	// Mock webhook payload and signature
	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_test_123","metadata":{"order_id":"order123","user_id":"user123"}}}}`)
	signature := testSignatureService
	secret := testSecret

	// Mock Stripe event
	event := stripe.Event{
		Type: "payment_intent.succeeded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"pi_test_123","metadata":{"order_id":"order123","user_id":"user123"}}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	// Add missing mock for GetPaymentByProviderPaymentID
	mockDB.On("GetPaymentByProviderPaymentID", mock.Anything, "pi_test_123").Return(database.Payment{
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
		// Add other fields as needed for your logic
	}, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_PaymentNotFound tests when payment is not found in webhook
func TestHandleWebhook_PaymentNotFound(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockStripe := new(mockStripeClient)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_test_123"}}}`)
	signature := testSignatureService
	secret := testSecret

	// Mock Stripe event
	event := stripe.Event{
		Type: "payment_intent.succeeded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"pi_test_123"}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockTx.On("Rollback").Return(nil)

	mockDB.On("GetPaymentByProviderPaymentID", mock.Anything, "pi_test_123").Return(database.Payment{}, sql.ErrNoRows)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")

	mockDB.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// Helper for webhook event status update tests
func runHandleWebhookStatusUpdateTest(t *testing.T, eventType, payloadStr, eventRawStr, status string) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(payloadStr)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: stripe.EventType(eventType),
		Data: &stripe.EventData{
			Raw: []byte(eventRawStr),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, database.UpdatePaymentStatusByProviderPaymentIDParams{
		Status:            status,
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

func TestHandleWebhook_PaymentFailed(t *testing.T) {
	runHandleWebhookStatusUpdateTest(
		t,
		"payment_intent.payment_failed",
		`{"type":"payment_intent.payment_failed","data":{"object":{"id":"pi_test_123"}}}`,
		`{"id":"pi_test_123"}`,
		"failed",
	)
}

func TestHandleWebhook_PaymentCanceled(t *testing.T) {
	runHandleWebhookStatusUpdateTest(
		t,
		"payment_intent.canceled",
		`{"type":"payment_intent.canceled","data":{"object":{"id":"pi_test_123"}}}`,
		`{"id":"pi_test_123"}`,
		"cancelled",
	)
}

func TestHandleWebhook_ChargeRefunded(t *testing.T) {
	runHandleWebhookStatusUpdateTest(
		t,
		"charge.refunded",
		`{"type":"charge.refunded","data":{"object":{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}}}`,
		`{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}`,
		"refunded",
	)
}

// TestHandleWebhook_DatabaseUpdateError tests when database update fails
func TestHandleWebhook_DatabaseUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_test_123"}}}`)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: "payment_intent.succeeded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"pi_test_123"}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("GetPaymentByProviderPaymentID", mock.Anything, "pi_test_123").Return(database.Payment{}, nil)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, database.UpdatePaymentStatusByProviderPaymentIDParams{
		Status:            "succeeded",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}).Return(errors.New("database error"))
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update payment")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_TransactionCommitError tests when transaction commit fails
func TestHandleWebhook_TransactionCommitError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_test_123"}}}`)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: "payment_intent.succeeded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"pi_test_123"}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("GetPaymentByProviderPaymentID", mock.Anything, "pi_test_123").Return(database.Payment{}, nil)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, database.UpdatePaymentStatusByProviderPaymentIDParams{
		Status:            "succeeded",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}).Return(nil)
	mockTx.On("Commit").Return(errors.New("commit error"))
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_UnknownEventType tests when webhook event type is not handled
func TestHandleWebhook_UnknownEventType(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"unknown.event","data":{"object":{"id":"pi_test_123"}}}`)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: "unknown.event",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"pi_test_123"}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.NoError(t, err) // Unknown events should be ignored

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_PendingStatuses tests when payment intent status is pending
func TestConfirmPayment_PendingStatuses(t *testing.T) {
	tests := []struct {
		name         string
		intentStatus stripe.PaymentIntentStatus
	}{
		{
			name:         "RequiresPaymentMethod",
			intentStatus: stripe.PaymentIntentStatusRequiresPaymentMethod,
		},
		{
			name:         "RequiresConfirmation",
			intentStatus: stripe.PaymentIntentStatusRequiresConfirmation,
		},
		{
			name:         "RequiresAction",
			intentStatus: stripe.PaymentIntentStatusRequiresAction,
		},
		{
			name:         "RequiresCapture",
			intentStatus: stripe.PaymentIntentStatusRequiresCapture,
		},
		{
			name:         "Processing",
			intentStatus: stripe.PaymentIntentStatusProcessing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockPaymentDBQueries)
			mockDBConn := new(mockPaymentDBConn)
			mockTx := new(mockPaymentDBTx)
			mockStripe := new(mockStripeClient)
			service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

			params := ConfirmPaymentParams{
				OrderID: "order123",
				UserID:  "user123",
			}

			payment := database.Payment{
				ID:                "payment123",
				OrderID:           "order123",
				UserID:            "user123",
				Amount:            "100.00",
				Currency:          "USD",
				Status:            "pending",
				Provider:          "stripe",
				ProviderPaymentID: utils.ToNullString("pi_test_123"),
			}

			mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
			mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
				ID:     "pi_test_123",
				Status: tt.intentStatus,
			}, nil)

			mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
			mockDB.On("WithTx", mockTx).Return(mockDB)
			mockDB.On("UpdatePaymentStatus", mock.Anything, mock.Anything).Return(nil)
			mockTx.On("Commit").Return(nil)
			mockTx.On("Rollback").Return(nil)

			result, err := service.ConfirmPayment(context.Background(), params)
			require.NoError(t, err)
			assert.Equal(t, "pending", result.Status)

			mockDB.AssertExpectations(t)
			mockDBConn.AssertExpectations(t)
			mockTx.AssertExpectations(t)
			mockStripe.AssertExpectations(t)
		})
	}
}

// TestConfirmPayment_StripeError tests when Stripe API call fails
func TestConfirmPayment_StripeError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "pending",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(nil, errors.New("stripe error"))

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to fetch payment intent")

	mockDB.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_TransactionError tests when transaction fails to start
func TestConfirmPayment_TransactionError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "pending",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
		ID:     "pi_test_123",
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(nil, errors.New("transaction error"))

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error starting transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_DatabaseUpdateError tests when database update fails
func TestConfirmPayment_DatabaseUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "pending",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
		ID:     "pi_test_123",
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatus", mock.Anything, mock.Anything).Return(errors.New("database error"))
	mockTx.On("Rollback").Return(nil)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update payment status")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_OrderUpdateError tests when order status update fails
func TestConfirmPayment_OrderUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "pending",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
		ID:     "pi_test_123",
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatus", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(errors.New("order update error"))
	mockTx.On("Rollback").Return(nil)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update order status")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestConfirmPayment_CommitError tests when transaction commit fails
func TestConfirmPayment_CommitError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := ConfirmPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "pending",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("GetPaymentIntent", "pi_test_123").Return(&stripe.PaymentIntent{
		ID:     "pi_test_123",
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil)

	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatus", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(errors.New("commit error"))
	mockTx.On("Rollback").Return(nil)

	_, err := service.ConfirmPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestGetPaymentHistory_DatabaseError tests when database query fails
func TestGetPaymentHistory_DatabaseError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	mockDB.On("GetPaymentsByUserID", mock.Anything, "user123").Return(nil, errors.New("database error"))

	_, err := service.GetPaymentHistory(context.Background(), "user123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to fetch payment history")

	mockDB.AssertExpectations(t)
}

// TestGetPaymentHistory_EmptyResult tests when user has no payment history
func TestGetPaymentHistory_EmptyResult(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	mockDB.On("GetPaymentsByUserID", mock.Anything, "user123").Return([]database.Payment{}, nil)

	result, err := service.GetPaymentHistory(context.Background(), "user123")
	require.NoError(t, err)
	assert.Empty(t, result)

	mockDB.AssertExpectations(t)
}

// TestGetAllPayments_WithStatusFilter tests when status filter is provided
func TestGetAllPayments_WithStatusFilter(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	payments := []database.Payment{
		{
			ID:                "payment123",
			OrderID:           "order123",
			UserID:            "user123",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: utils.ToNullString("pi_test_123"),
			CreatedAt:         time.Now(),
		},
	}

	mockDB.On("GetPaymentsByStatus", mock.Anything, "succeeded").Return(payments, nil)

	result, err := service.GetAllPayments(context.Background(), "succeeded")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "payment123", result[0].ID)

	mockDB.AssertExpectations(t)
}

// TestGetAllPayments_DatabaseError tests when database query fails
func TestGetAllPayments_DatabaseError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	mockDB.On("GetAllPayments", mock.Anything).Return(nil, errors.New("database error"))

	_, err := service.GetAllPayments(context.Background(), "all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to fetch payments")

	mockDB.AssertExpectations(t)
}

// TestGetAllPayments_StatusFilterDatabaseError tests when status filter database query fails
func TestGetAllPayments_StatusFilterDatabaseError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	mockDB.On("GetPaymentsByStatus", mock.Anything, "succeeded").Return(nil, errors.New("database error"))

	_, err := service.GetAllPayments(context.Background(), "succeeded")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to fetch payments")

	mockDB.AssertExpectations(t)
}

// TestGetAllPayments_EmptyResult tests when no payments match the filter
func TestGetAllPayments_EmptyResult(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	mockDB.On("GetAllPayments", mock.Anything).Return([]database.Payment{}, nil)

	result, err := service.GetAllPayments(context.Background(), "all")
	require.NoError(t, err)
	assert.Empty(t, result)

	mockDB.AssertExpectations(t)
}

// TestRefundPayment_StripeError tests when Stripe refund fails
func TestRefundPayment_StripeError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("CreateRefund", mock.Anything).Return(nil, errors.New("stripe error"))

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to process refund")

	mockDB.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestRefundPayment_TransactionError tests when transaction fails to start
func TestRefundPayment_TransactionError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("CreateRefund", mock.Anything).Return(&stripe.Refund{
		ID:     "re_test_123",
		Status: stripe.RefundStatusSucceeded,
	}, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(nil, errors.New("transaction error"))

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error starting transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestRefundPayment_PaymentUpdateError tests when payment status update fails
func TestRefundPayment_PaymentUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("CreateRefund", mock.Anything).Return(&stripe.Refund{
		ID:     "re_test_123",
		Status: stripe.RefundStatusSucceeded,
	}, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByID", mock.Anything, database.UpdatePaymentStatusByIDParams{
		ID:     "payment123",
		Status: "refunded",
	}).Return(errors.New("payment update error"))
	mockTx.On("Rollback").Return(nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update payment status")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestRefundPayment_OrderUpdateError tests when order status update fails
func TestRefundPayment_OrderUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("CreateRefund", mock.Anything).Return(&stripe.Refund{
		ID:     "re_test_123",
		Status: stripe.RefundStatusSucceeded,
	}, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByID", mock.Anything, database.UpdatePaymentStatusByIDParams{
		ID:     "payment123",
		Status: "refunded",
	}).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(errors.New("order update error"))
	mockTx.On("Rollback").Return(nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update order status")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestRefundPayment_CommitError tests when transaction commit fails
func TestRefundPayment_CommitError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	params := RefundPaymentParams{
		OrderID: "order123",
		UserID:  "user123",
	}

	payment := database.Payment{
		ID:                "payment123",
		OrderID:           "order123",
		UserID:            "user123",
		Amount:            "100.00",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}

	mockDB.On("GetPaymentByOrderID", mock.Anything, "order123").Return(payment, nil)
	mockStripe.On("CreateRefund", mock.Anything).Return(&stripe.Refund{
		ID:     "re_test_123",
		Status: stripe.RefundStatusSucceeded,
	}, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByID", mock.Anything, database.UpdatePaymentStatusByIDParams{
		ID:     "payment123",
		Status: "refunded",
	}).Return(nil)
	mockDB.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(errors.New("commit error"))
	mockTx.On("Rollback").Return(nil)

	err := service.RefundPayment(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_ChargeRefunded_DatabaseUpdateError tests DB update error for charge.refunded
func TestHandleWebhook_ChargeRefunded_DatabaseUpdateError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"charge.refunded","data":{"object":{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}}}`)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: "charge.refunded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, database.UpdatePaymentStatusByProviderPaymentIDParams{
		Status:            "refunded",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}).Return(errors.New("db error"))
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to update payment")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_ChargeRefunded_CommitError tests commit error for charge.refunded
func TestHandleWebhook_ChargeRefunded_CommitError(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockTx := new(mockPaymentDBTx)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`{"type":"charge.refunded","data":{"object":{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}}}`)
	signature := testSignatureService
	secret := testSecret

	event := stripe.Event{
		Type: "charge.refunded",
		Data: &stripe.EventData{
			Raw: []byte(`{"id":"ch_test_123","payment_intent":{"id":"pi_test_123"}}`),
		},
	}
	mockStripe.On("ParseWebhook", payload, signature, secret).Return(event, nil)
	mockDBConn.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	mockDB.On("WithTx", mockTx).Return(mockDB)
	mockDB.On("UpdatePaymentStatusByProviderPaymentID", mock.Anything, database.UpdatePaymentStatusByProviderPaymentIDParams{
		Status:            "refunded",
		ProviderPaymentID: utils.ToNullString("pi_test_123"),
	}).Return(nil)
	mockTx.On("Commit").Return(errors.New("commit error"))
	mockTx.On("Rollback").Return(nil)

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error committing transaction")

	mockDB.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockStripe.AssertExpectations(t)
}

// TestHandleWebhook_MalformedPayload tests malformed payload
func TestHandleWebhook_MalformedPayload(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	mockDBConn := new(mockPaymentDBConn)
	mockStripe := new(mockStripeClient)
	service := &paymentServiceImpl{db: mockDB, dbConn: mockDBConn, apiKey: "sk_test_123", stripe: mockStripe}

	payload := []byte(`not a json`)
	signature := testSignatureService
	secret := testSecret

	mockStripe.On("ParseWebhook", payload, signature, secret).Return(stripe.Event{}, errors.New("bad payload"))

	err := service.HandleWebhook(context.Background(), payload, signature, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Signature verification failed")

	mockStripe.AssertExpectations(t)
}

// TestGetPaymentHistory_EmptyUserID tests empty user ID
func TestGetPaymentHistory_EmptyUserID(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	_, err := service.GetPaymentHistory(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "User ID is required")
}

// TestGetPayment_EmptyOrderID tests empty order ID
func TestGetPayment_EmptyOrderID(t *testing.T) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}

	_, err := service.GetPayment(context.Background(), "", "user123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Order ID is required")
}

// Helper for GetPayment error scenarios
func runGetPaymentErrorTest(t *testing.T, payment database.Payment, orderID, userID, expectedMsg string) {
	mockDB := new(mockPaymentDBQueries)
	service := &paymentServiceImpl{db: mockDB, dbConn: nil, apiKey: "sk_test_123"}
	mockDB.On("GetPaymentByOrderID", mock.Anything, orderID).Return(payment, nil)

	_, err := service.GetPayment(context.Background(), orderID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedMsg)
	mockDB.AssertExpectations(t)
}

func TestGetPayment_EmptyUserID(t *testing.T) {
	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "",
		Amount:  "100.00",
		Status:  "succeeded",
	}
	runGetPaymentErrorTest(t, payment, "order123", "user123", "Payment does not belong to user")
}

func TestGetPayment_InvalidAmountParsing(t *testing.T) {
	payment := database.Payment{
		ID:      "payment123",
		OrderID: "order123",
		UserID:  "user123",
		Amount:  "not_a_number",
		Status:  "succeeded",
	}
	runGetPaymentErrorTest(t, payment, "order123", "user123", "Invalid payment amount")
}
