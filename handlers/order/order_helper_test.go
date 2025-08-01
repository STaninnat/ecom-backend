// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// order_helper_test.go: Provides mock implementations of database queries, transaction, service, and logger interfaces for unit testing.

// MockDBQueries is a mock implementation of database queries for testing
type MockDBQueries struct {
	mock.Mock
}

func (m *MockDBQueries) CreateOrder(ctx context.Context, params database.CreateOrderParams) (database.Order, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(database.Order), args.Error(1)
}

func (m *MockDBQueries) CreateOrderItem(ctx context.Context, params database.CreateOrderItemParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) ListAllOrders(ctx context.Context) ([]database.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Order), args.Error(1)
}

func (m *MockDBQueries) GetOrderByUserID(ctx context.Context, userID string) ([]database.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]database.Order), args.Error(1)
}

func (m *MockDBQueries) GetOrderByID(ctx context.Context, orderID string) (database.Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(database.Order), args.Error(1)
}

func (m *MockDBQueries) GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]database.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]database.OrderItem), args.Error(1)
}

func (m *MockDBQueries) UpdateOrderStatus(ctx context.Context, params database.UpdateOrderStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) DeleteOrderByID(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockDBQueries) WithTx(tx *sql.Tx) *MockDBQueries {
	args := m.Called(tx)
	return args.Get(0).(*MockDBQueries)
}

// MockDBConn is a mock implementation of database connection for testing
type MockDBConn struct {
	mock.Mock
}

func (m *MockDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// MockTx is a mock implementation of database transaction for testing
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockOrderService is a mock implementation of OrderService for testing
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, user database.User, params CreateOrderRequest) (*OrderResponse, error) {
	args := m.Called(ctx, user, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderResponse), args.Error(1)
}

func (m *MockOrderService) GetAllOrders(ctx context.Context) ([]database.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Order), args.Error(1)
}

func (m *MockOrderService) GetUserOrders(ctx context.Context, user database.User) ([]UserOrderResponse, error) {
	args := m.Called(ctx, user)
	return args.Get(0).([]UserOrderResponse), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, orderID string, user database.User) (*OrderDetailResponse, error) {
	args := m.Called(ctx, orderID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderDetailResponse), args.Error(1)
}

func (m *MockOrderService) GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]OrderItemResponse, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]OrderItemResponse), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrderService) DeleteOrder(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

// MockHandlersConfig is a mock implementation of HandlersConfig for testing
type MockHandlersConfig struct {
	mock.Mock
}

func (m *MockHandlersConfig) LogHandlerError(ctx context.Context, operation, code, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, code, message, ip, userAgent, err)
}

func (m *MockHandlersConfig) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}

// mockHandlerLogger is a mock implementation of HandlerLogger interface for testing
type mockHandlerLogger struct {
	mock.Mock
}

func (m *mockHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// testHandlerCreateOrderError tests order creation handler for expected error responses.
func testHandlerCreateOrderError(t *testing.T, user database.User, requestBody CreateOrderRequest, expectedCode int, expectedMessage string, expectedAppError *handlers.AppError) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, expectedAppError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", expectedAppError.Code, expectedAppError.Message, mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, expectedCode, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedMessage, response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// testHandlerCreateOrderServerError tests order creation handler for server error responses.
func testHandlerCreateOrderServerError(t *testing.T, user database.User, requestBody CreateOrderRequest, expectedCode int, expectedAppError *handlers.AppError) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	mockOrderService.On("CreateOrder", mock.Anything, user, requestBody).Return(nil, expectedAppError)
	mockLogger.On("LogHandlerError", mock.Anything, "create_order", expectedAppError.Code, expectedAppError.Message, mock.Anything, mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerCreateOrder(w, req, user)

	assert.Equal(t, expectedCode, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// For server error, a general message will be sent.
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}
