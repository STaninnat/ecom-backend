package carthandlers

import (
	"context"
	"database/sql"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/mock"
)

type MockCartService struct{ mock.Mock }

func (m *MockCartService) AddItemToUserCart(ctx context.Context, userID, productID string, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}
func (m *MockCartService) AddItemToGuestCart(ctx context.Context, sessionID, productID string, quantity int) error {
	args := m.Called(ctx, sessionID, productID, quantity)
	return args.Error(0)
}
func (m *MockCartService) GetUserCart(ctx context.Context, userID string) (*models.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}
func (m *MockCartService) GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}
func (m *MockCartService) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}
func (m *MockCartService) UpdateGuestItemQuantity(ctx context.Context, sessionID, productID string, quantity int) error {
	args := m.Called(ctx, sessionID, productID, quantity)
	return args.Error(0)
}
func (m *MockCartService) RemoveItem(ctx context.Context, userID, productID string) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}
func (m *MockCartService) RemoveGuestItem(ctx context.Context, sessionID, productID string) error {
	args := m.Called(ctx, sessionID, productID)
	return args.Error(0)
}
func (m *MockCartService) DeleteUserCart(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockCartService) DeleteGuestCart(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}
func (m *MockCartService) CheckoutUserCart(ctx context.Context, userID string) (*CartCheckoutResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CartCheckoutResult), args.Error(1)
}
func (m *MockCartService) CheckoutGuestCart(ctx context.Context, sessionID, userID string) (*CartCheckoutResult, error) {
	args := m.Called(ctx, sessionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CartCheckoutResult), args.Error(1)
}

type MockLogger struct{ mock.Mock }

func (m *MockLogger) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}
func (m *MockLogger) LogHandlerError(ctx context.Context, operation, code, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, code, message, ip, userAgent, err)
}

// MockCartMongoAPI is a mock implementation of CartMongoAPI for testing
type MockCartMongoAPI struct {
	mock.Mock
}

func (m *MockCartMongoAPI) AddItemToCart(ctx context.Context, userID string, item models.CartItem) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func (m *MockCartMongoAPI) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartMongoAPI) UpdateItemQuantity(ctx context.Context, userID string, productID string, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}

func (m *MockCartMongoAPI) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartMongoAPI) ClearCart(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockProductAPI is a mock implementation of ProductAPI for testing
type MockProductAPI struct {
	mock.Mock
}

func (m *MockProductAPI) GetProductByID(ctx context.Context, productID string) (database.Product, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(database.Product), args.Error(1)
}

func (m *MockProductAPI) UpdateProductStock(ctx context.Context, params database.UpdateProductStockParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// MockOrderAPI is a mock implementation of OrderAPI for testing
type MockOrderAPI struct {
	mock.Mock
}

func (m *MockOrderAPI) CreateOrder(ctx context.Context, params database.CreateOrderParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockOrderAPI) CreateOrderItem(ctx context.Context, params database.CreateOrderItemParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// MockDBConnAPI is a mock implementation of DBConnAPI for testing
type MockDBConnAPI struct {
	mock.Mock
}

func (m *MockDBConnAPI) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTxAPI, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(DBTxAPI), args.Error(1)
}

// MockDBTxAPI is a mock implementation of DBTxAPI for testing
type MockDBTxAPI struct {
	mock.Mock
}

func (m *MockDBTxAPI) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBTxAPI) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockCartRedisAPI is a mock implementation of CartRedisAPI for testing
type MockCartRedisAPI struct {
	mock.Mock
}

func (m *MockCartRedisAPI) GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartRedisAPI) SaveGuestCart(ctx context.Context, sessionID string, cart *models.Cart) error {
	args := m.Called(ctx, sessionID, cart)
	return args.Error(0)
}

func (m *MockCartRedisAPI) UpdateGuestItemQuantity(ctx context.Context, sessionID, productID string, quantity int) error {
	args := m.Called(ctx, sessionID, productID, quantity)
	return args.Error(0)
}

func (m *MockCartRedisAPI) RemoveGuestItem(ctx context.Context, sessionID, productID string) error {
	args := m.Called(ctx, sessionID, productID)
	return args.Error(0)
}

func (m *MockCartRedisAPI) DeleteGuestCart(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}
