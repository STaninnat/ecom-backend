package paymenthandlers

import (
	"context"
	"database/sql"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v82"
)

// MockPaymentService is a testify-based mock implementation of the PaymentService interface.
// It allows tests to set up expected method calls and return values for testing handlers without a real service.
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CreatePaymentResult), args.Error(1)
}

func (m *MockPaymentService) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConfirmPaymentResult), args.Error(1)
}

func (m *MockPaymentService) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	args := m.Called(ctx, orderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GetPaymentResult), args.Error(1)
}

func (m *MockPaymentService) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PaymentHistoryItem), args.Error(1)
}

func (m *MockPaymentService) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PaymentHistoryItem), args.Error(1)
}

func (m *MockPaymentService) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockPaymentService) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	args := m.Called(ctx, payload, signature, secret)
	return args.Error(0)
}

// --- Service Mock ---
// mockPaymentDBQueries is a testify-based mock implementation of PaymentDBQueries.
// It allows tests to mock database query operations without a real database.
type mockPaymentDBQueries struct{ mock.Mock }

func (m *mockPaymentDBQueries) WithTx(tx PaymentDBTx) PaymentDBQueries {
	args := m.Called(tx)
	if q, ok := args.Get(0).(PaymentDBQueries); ok {
		return q
	}
	return m
}

func (m *mockPaymentDBQueries) GetOrderByID(ctx context.Context, id string) (database.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Order), args.Error(1)
}

func (m *mockPaymentDBQueries) GetPaymentByOrderID(ctx context.Context, orderID string) (database.Payment, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(database.Payment), args.Error(1)
}

func (m *mockPaymentDBQueries) GetPaymentsByUserID(ctx context.Context, userID string) ([]database.Payment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Payment), args.Error(1)
}

func (m *mockPaymentDBQueries) GetAllPayments(ctx context.Context) ([]database.Payment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Payment), args.Error(1)
}

func (m *mockPaymentDBQueries) GetPaymentsByStatus(ctx context.Context, status string) ([]database.Payment, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Payment), args.Error(1)
}

func (m *mockPaymentDBQueries) CreatePayment(ctx context.Context, params database.CreatePaymentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockPaymentDBQueries) UpdatePaymentStatus(ctx context.Context, params database.UpdatePaymentStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockPaymentDBQueries) UpdatePaymentStatusByID(ctx context.Context, params database.UpdatePaymentStatusByIDParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockPaymentDBQueries) UpdatePaymentStatusByProviderPaymentID(ctx context.Context, params database.UpdatePaymentStatusByProviderPaymentIDParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockPaymentDBQueries) UpdateOrderStatus(ctx context.Context, params database.UpdateOrderStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockPaymentDBQueries) GetPaymentByProviderPaymentID(ctx context.Context, providerPaymentID string) (database.Payment, error) {
	args := m.Called(ctx, providerPaymentID)
	return args.Get(0).(database.Payment), args.Error(1)
}

// --- Database Connection Mock ---
// mockPaymentDBConn is a testify-based mock implementation of PaymentDBConn.
type mockPaymentDBConn struct{ mock.Mock }

func (m *mockPaymentDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (PaymentDBTx, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(PaymentDBTx), args.Error(1)
}

// --- Database Transaction Mock ---
// mockPaymentDBTx is a testify-based mock implementation of PaymentDBTx.
type mockPaymentDBTx struct{ mock.Mock }

func (m *mockPaymentDBTx) Rollback() error { return m.Called().Error(0) }
func (m *mockPaymentDBTx) Commit() error   { return m.Called().Error(0) }

// mockStripeClient is a testify-based mock implementation of StripeClient for testing.
type mockStripeClient struct{ mock.Mock }

func (m *mockStripeClient) CreatePaymentIntent(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}
func (m *mockStripeClient) GetPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}
func (m *mockStripeClient) CreateRefund(params *stripe.RefundParams) (*stripe.Refund, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Refund), args.Error(1)
}
func (m *mockStripeClient) ParseWebhook(payload []byte, sigHeader, secret string) (stripe.Event, error) {
	args := m.Called(payload, sigHeader, secret)
	if args.Get(0) == nil {
		return stripe.Event{}, args.Error(1)
	}
	return args.Get(0).(stripe.Event), args.Error(1)
}

// MockPaymentServiceForConfirm is a mock implementation of PaymentService for confirm tests
// Only ConfirmPayment is implemented; others are stubs
type MockPaymentServiceForConfirm struct {
	mock.Mock
}

func (m *MockPaymentServiceForConfirm) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForConfirm) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConfirmPaymentResult), args.Error(1)
}
func (m *MockPaymentServiceForConfirm) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForConfirm) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForConfirm) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForConfirm) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	return nil
}
func (m *MockPaymentServiceForConfirm) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	return nil
}

// MockLoggerForConfirm is a mock implementation of HandlerLogger for confirm tests
type MockLoggerForConfirm struct {
	mock.Mock
}

func (m *MockLoggerForConfirm) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *MockLoggerForConfirm) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockPaymentServiceForCreate is a mock implementation of PaymentService
// specifically for testing payment creation functionality
type MockPaymentServiceForCreate struct {
	mock.Mock
}

func (m *MockPaymentServiceForCreate) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CreatePaymentResult), args.Error(1)
}

func (m *MockPaymentServiceForCreate) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	return nil, nil // not used in create tests
}

func (m *MockPaymentServiceForCreate) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	return nil, nil // not used in create tests
}

func (m *MockPaymentServiceForCreate) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	return nil, nil // not used in create tests
}

func (m *MockPaymentServiceForCreate) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	return nil, nil // not used in create tests
}

func (m *MockPaymentServiceForCreate) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	return nil // not used in create tests
}

func (m *MockPaymentServiceForCreate) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	return nil // not used in create tests
}

// MockLoggerForCreate is a mock implementation of HandlerLogger
// specifically for testing payment creation logging
type MockLoggerForCreate struct {
	mock.Mock
}

func (m *MockLoggerForCreate) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *MockLoggerForCreate) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockPaymentServiceForGet is a mock implementation of PaymentService for get tests
// Only GetPayment is implemented; others are stubs
type MockPaymentServiceForGet struct {
	mock.Mock
}

func (m *MockPaymentServiceForGet) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForGet) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForGet) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	args := m.Called(ctx, orderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GetPaymentResult), args.Error(1)
}
func (m *MockPaymentServiceForGet) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PaymentHistoryItem), args.Error(1)
}
func (m *MockPaymentServiceForGet) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PaymentHistoryItem), args.Error(1)
}
func (m *MockPaymentServiceForGet) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	return nil
}
func (m *MockPaymentServiceForGet) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	return nil
}

// MockLoggerForGet is a mock implementation of HandlerLogger for get tests
type MockLoggerForGet struct {
	mock.Mock
}

func (m *MockLoggerForGet) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *MockLoggerForGet) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockPaymentServiceForRefund is a mock implementation of PaymentService for refund tests
// Only RefundPayment is implemented; others are stubs
type MockPaymentServiceForRefund struct {
	mock.Mock
}

func (m *MockPaymentServiceForRefund) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForRefund) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForRefund) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForRefund) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForRefund) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForRefund) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}
func (m *MockPaymentServiceForRefund) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	return nil
}

// MockLoggerForRefund is a mock implementation of HandlerLogger for refund tests
type MockLoggerForRefund struct {
	mock.Mock
}

func (m *MockLoggerForRefund) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *MockLoggerForRefund) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockPaymentServiceForWebhook is a mock implementation of PaymentService for webhook tests
// Only HandleWebhook is implemented; others are stubs
type MockPaymentServiceForWebhook struct {
	mock.Mock
}

func (m *MockPaymentServiceForWebhook) CreatePayment(ctx context.Context, params CreatePaymentParams) (*CreatePaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForWebhook) ConfirmPayment(ctx context.Context, params ConfirmPaymentParams) (*ConfirmPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForWebhook) GetPayment(ctx context.Context, orderID string, userID string) (*GetPaymentResult, error) {
	return nil, nil
}
func (m *MockPaymentServiceForWebhook) GetPaymentHistory(ctx context.Context, userID string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForWebhook) GetAllPayments(ctx context.Context, status string) ([]PaymentHistoryItem, error) {
	return nil, nil
}
func (m *MockPaymentServiceForWebhook) RefundPayment(ctx context.Context, params RefundPaymentParams) error {
	return nil
}
func (m *MockPaymentServiceForWebhook) HandleWebhook(ctx context.Context, payload []byte, signature string, secret string) error {
	args := m.Called(ctx, payload, signature, secret)
	return args.Error(0)
}

// MockLoggerForWebhook is a mock implementation of HandlerLogger for webhook tests
type MockLoggerForWebhook struct {
	mock.Mock
}

func (m *MockLoggerForWebhook) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *MockLoggerForWebhook) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockHandlersConfig is a mock implementation of HandlerLogger for testing
type MockHandlersConfig struct {
	mock.Mock
}

func (m *MockHandlersConfig) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *MockHandlersConfig) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}
