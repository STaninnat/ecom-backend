// Package producthandlers provides HTTP handlers and business logic for managing products, including CRUD operations and filtering.
package producthandlers

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// product_helper_test.go: Provides testify-based mocks for ProductService, Logger, and database queries to support unit testing without real dependencies.

// MockProductService is a testify-based mock implementation of the ProductService interface.
// It allows tests to set up expected method calls and return values for testing handlers without a real service.
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, params ProductRequest) (string, error) {
	args := m.Called(ctx, params)
	return args.String(0), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, params ProductRequest) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductService) GetAllProducts(ctx context.Context, isAdmin bool) ([]database.Product, error) {
	args := m.Called(ctx, isAdmin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, productID string, isAdmin bool) (database.Product, error) {
	args := m.Called(ctx, productID, isAdmin)
	return args.Get(0).(database.Product), args.Error(1)
}

func (m *MockProductService) FilterProducts(ctx context.Context, params FilterProductsRequest) ([]database.Product, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Product), args.Error(1)
}

// --- Mock Logger ---
// mockLogger is a testify-based mock implementation of the Logger interface.
// It allows tests to verify that logging methods are called with expected parameters.
type mockLogger struct{ mock.Mock }

func (m *mockLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *mockLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// --- Service Mock ---
// mockDBQueries is a testify-based mock implementation of database.Queries.
// It allows tests to mock database query operations without a real database.
type mockDBQueries struct{ mock.Mock }
type mockDBConn struct{ mock.Mock }
type mockTx struct{ mock.Mock }

func (m *mockDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (ProductDBTx, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(ProductDBTx), args.Error(1)
}
func (m *mockTx) Commit() error   { return m.Called().Error(0) }
func (m *mockTx) Rollback() error { return m.Called().Error(0) }

func (m *mockDBQueries) WithTx(tx ProductDBTx) ProductDBQueries {
	args := m.Called(tx)
	if q, ok := args.Get(0).(ProductDBQueries); ok {
		return q
	}
	return m
}
func (m *mockDBQueries) CreateProduct(ctx context.Context, params database.CreateProductParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}
func (m *mockDBQueries) UpdateProduct(ctx context.Context, params database.UpdateProductParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}
func (m *mockDBQueries) DeleteProductByID(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockDBQueries) GetAllProducts(ctx context.Context) ([]database.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Product), args.Error(1)
}
func (m *mockDBQueries) GetAllActiveProducts(ctx context.Context) ([]database.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Product), args.Error(1)
}
func (m *mockDBQueries) GetProductByID(ctx context.Context, id string) (database.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Product), args.Error(1)
}
func (m *mockDBQueries) GetActiveProductByID(ctx context.Context, id string) (database.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Product), args.Error(1)
}
func (m *mockDBQueries) FilterProducts(ctx context.Context, params database.FilterProductsParams) ([]database.Product, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]database.Product), args.Error(1)
}
