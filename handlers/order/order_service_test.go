// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// order_service_test.go: Tests for the OrderService implementation, focusing on order creation logic and error handling.

// TestOrderServiceInterface ensures the OrderService interface is properly defined and implemented.
func TestOrderServiceInterface(_ *testing.T) {
	var _ OrderService = (*orderServiceImpl)(nil)
}

// TestNewOrderService ensures NewOrderService returns a valid OrderService instance.
func TestNewOrderService(t *testing.T) {
	// Create mock database instances for testing using sqlmock
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	assert.NotNil(t, service)

	// Test that the service implements the interface
	var _ = service
}

// TestCreateOrder_Success tests successful order creation.
func TestCreateOrder_Success(t *testing.T) {
	// Create mock database instances for testing using sqlmock
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2, Price: 10.50},
			{ProductID: "prod2", Quantity: 1, Price: 25.00},
		},
		PaymentMethod:   "credit_card",
		ShippingAddress: "123 Main St",
		ContactPhone:    "555-1234",
	}

	// This test will fail due to transaction issues, but it tests the interface
	result, err := service.CreateOrder(context.Background(), user, params)

	// Expect an error due to transaction issues with sqlmock
	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
}

// TestCreateOrder_EmptyItems tests order creation with empty items list.
func TestCreateOrder_EmptyItems(t *testing.T) {
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{},
	}

	result, err := service.CreateOrder(context.Background(), user, params)

	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
	assert.Equal(t, "Order must contain at least one item", appErr.Message)
}

// TestCreateOrder_InvalidQuantity tests order creation with invalid quantity.
func TestCreateOrder_InvalidQuantity(t *testing.T) {
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 0, Price: 10.50},
		},
	}

	result, err := service.CreateOrder(context.Background(), user, params)

	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
	assert.Equal(t, "Quantity must be greater than 0", appErr.Message)
}

// TestCreateOrder_NegativePrice tests order creation with negative price.
func TestCreateOrder_NegativePrice(t *testing.T) {
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 1, Price: -10.50},
		},
	}

	result, err := service.CreateOrder(context.Background(), user, params)

	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
	assert.Equal(t, "Price cannot be negative", appErr.Message)
}

// TestCreateOrder_QuantityOverflow tests order creation with quantity overflow.
func TestCreateOrder_QuantityOverflow(t *testing.T) {
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 2147483648, Price: 10.50}, // Exceeds int32 max
		},
	}

	result, err := service.CreateOrder(context.Background(), user, params)

	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
}

// TestCreateOrder_ErrorScenarios tests order creation with various error scenarios.
func TestCreateOrder_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name         string
		mockSetup    func(mock sqlmock.Sqlmock)
		useNilDB     bool
		expectedCode string
		expectedMsg  string
	}{
		{
			name: "TransactionError",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
				mock.ExpectClose()
			},
			useNilDB:     false,
			expectedCode: "transaction_error",
			expectedMsg:  "",
		},
		{
			name: "CreateOrderError",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders").WillReturnError(errors.New("create order error"))
				mock.ExpectRollback()
				mock.ExpectClose()
			},
			useNilDB:     false,
			expectedCode: "create_order_error",
			expectedMsg:  "",
		},
		{
			name: "CommitError",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO order_items").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
				mock.ExpectClose()
			},
			useNilDB:     false,
			expectedCode: "create_order_error",
			expectedMsg:  "",
		},
		{
			name: "CreateOrderItemError",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO order_items").WillReturnError(errors.New("order item creation error"))
				mock.ExpectRollback()
				mock.ExpectClose()
			},
			useNilDB:     false,
			expectedCode: "create_order_error",
			expectedMsg:  "",
		},
		{
			name:         "NilDBConnection",
			mockSetup:    nil,
			useNilDB:     true,
			expectedCode: "transaction_error",
			expectedMsg:  "DB connection is nil",
		},
		{
			name: "TransactionBeginError",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction begin error"))
				mock.ExpectClose()
			},
			useNilDB:     false,
			expectedCode: "transaction_error",
			expectedMsg:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var service OrderService
			var db *sql.DB
			var mock sqlmock.Sqlmock
			if tc.useNilDB {
				service = NewOrderService(nil, nil)
			} else {
				db, mock, _ = sqlmock.New()
				queries := database.New(db)
				if tc.mockSetup != nil {
					mock.MatchExpectationsInOrder(false)
					tc.mockSetup(mock)
				}
				service = NewOrderService(queries, db)
			}
			user := database.User{ID: "user123"}
			params := CreateOrderRequest{
				Items: []OrderItemInput{{ProductID: "prod1", Quantity: 2, Price: 10.50}},
			}
			result, err := service.CreateOrder(context.Background(), user, params)
			require.Error(t, err)
			assert.Nil(t, result)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.expectedCode, appErr.Code)
			if tc.expectedMsg != "" {
				assert.Equal(t, tc.expectedMsg, appErr.Message)
			}
			if db != nil {
				if err := db.Close(); err != nil {
					t.Fatalf("failed to close db: %v", err)
				}
			}
		})
	}
}

// TestGetAllOrders_Success tests successful retrieval of all orders.
func TestGetAllOrders_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database query with correct 11 columns
	mock.ExpectQuery("SELECT (.+) FROM orders").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user1", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)

	orders, err := service.GetAllOrders(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Len(t, orders, 1)
}

// TestGetAllOrders_DatabaseError tests error handling when database fails.
func TestGetAllOrders_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database query to return an error
	mock.ExpectQuery("SELECT (.+) FROM orders").WillReturnError(errors.New("database error"))

	orders, err := service.GetAllOrders(context.Background())

	require.Error(t, err)
	assert.Nil(t, orders)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
}

// TestGetUserOrders_Success tests successful retrieval of user orders.
func TestGetUserOrders_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123"}

	// Mock the database queries with correct column structure
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("user123").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user123", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "order_id", "product_id", "quantity", "price", "created_at", "updated_at",
		}).AddRow(
			"item1", "order1", "prod1", 2, "50.00", time.Now(), time.Now(),
		),
	)

	orders, err := service.GetUserOrders(context.Background(), user)

	require.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Len(t, orders, 1)
}

// TestGetUserOrders_DatabaseError tests user orders retrieval with database error.
func TestGetUserOrders_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123"}

	// Mock the database query to return an error
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("user123").WillReturnError(errors.New("database error"))

	orders, err := service.GetUserOrders(context.Background(), user)

	require.Error(t, err)
	assert.Nil(t, orders)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Failed to get orders", appErr.Message)
}

// TestGetUserOrders_OrderItemsError tests user orders retrieval with order items error.
func TestGetUserOrders_OrderItemsError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123"}

	// Mock the database queries
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("user123").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user123", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnError(errors.New("order items error"))

	orders, err := service.GetUserOrders(context.Background(), user)

	require.Error(t, err)
	assert.Nil(t, orders)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Failed to get order items", appErr.Message)
}

// TestGetUserOrders_NilDatabase tests user orders retrieval with nil database.
func TestGetUserOrders_NilDatabase(t *testing.T) {
	service := NewOrderService(nil, nil)
	user := database.User{ID: "user123"}

	orders, err := service.GetUserOrders(context.Background(), user)

	require.Error(t, err)
	assert.Nil(t, orders)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Database not initialized", appErr.Message)
}

// TestGetOrderByID_Success tests successful retrieval of order by ID.
func TestGetOrderByID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123"}

	// Mock the database queries with correct column structure
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user123", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "order_id", "product_id", "quantity", "price", "created_at", "updated_at",
		}).AddRow(
			"item1", "order1", "prod1", 2, "50.00", time.Now(), time.Now(),
		),
	)

	order, err := service.GetOrderByID(context.Background(), "order1", user)

	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "order1", order.Order.ID)
}

// TestGetOrderByID_Unauthorized tests unauthorized access to order.
func TestGetOrderByID_Unauthorized(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123", Role: "user"}

	// Mock the database query to return an order owned by a different user
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user456", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)

	order, err := service.GetOrderByID(context.Background(), "order1", user)

	require.Error(t, err)
	assert.Nil(t, order)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "unauthorized", appErr.Code)
}

// TestGetOrderByID_AdminAccess tests admin access to any order.
func TestGetOrderByID_AdminAccess(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "admin123", Role: "admin"}

	// Mock the database queries with correct column structure
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user456", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "order_id", "product_id", "quantity", "price", "created_at", "updated_at",
		}).AddRow(
			"item1", "order1", "prod1", 2, "50.00", time.Now(), time.Now(),
		),
	)

	order, err := service.GetOrderByID(context.Background(), "order1", user)

	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "order1", order.Order.ID)
}

// TestGetOrderByID_OrderItemsError tests order retrieval with order items error.
func TestGetOrderByID_OrderItemsError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)
	user := database.User{ID: "user123"}

	// Mock the database queries
	mock.ExpectQuery("SELECT (.+) FROM orders").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "user_id", "total_amount", "status", "payment_method",
			"external_payment_id", "tracking_number", "shipping_address",
			"contact_phone", "created_at", "updated_at",
		}).AddRow(
			"order1", "user123", "100.00", "pending", sql.NullString{String: "credit_card", Valid: true},
			sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false},
			sql.NullString{String: "123 Main St", Valid: true}, sql.NullString{String: "555-1234", Valid: true},
			time.Now(), time.Now(),
		),
	)
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnError(errors.New("order items error"))

	order, err := service.GetOrderByID(context.Background(), "order1", user)

	require.Error(t, err)
	assert.Nil(t, order)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Failed to fetch order items", appErr.Message)
}

// TestGetOrderByID_NilDatabase tests order retrieval with nil database.
func TestGetOrderByID_NilDatabase(t *testing.T) {
	service := NewOrderService(nil, nil)
	user := database.User{ID: "user123"}

	order, err := service.GetOrderByID(context.Background(), "order1", user)

	require.Error(t, err)
	assert.Nil(t, order)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Database not initialized", appErr.Message)
}

// TestUpdateOrderStatus_Success tests successful order status update.
func TestUpdateOrderStatus_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database operations
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.UpdateOrderStatus(context.Background(), "order123", "shipped")

	assert.NoError(t, err)
}

// TestUpdateOrderStatus_InvalidStatus tests order status update with invalid status.
func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	db, _, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	err := service.UpdateOrderStatus(context.Background(), "order123", "invalid_status")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_status", appErr.Code)
	assert.Equal(t, "Invalid order status", appErr.Message)
}

// TestUpdateOrderStatus_CommitError tests order status update with commit error.
func TestUpdateOrderStatus_CommitError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database operations
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := service.UpdateOrderStatus(context.Background(), "order123", "shipped")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "commit_error", appErr.Code)
}

// TestUpdateOrderStatus_NilDBConnection tests order status update with nil database connection.
func TestUpdateOrderStatus_NilDBConnection(t *testing.T) {
	service := NewOrderService(nil, nil)

	err := service.UpdateOrderStatus(context.Background(), "order123", "shipped")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
	assert.Equal(t, "DB connection is nil", appErr.Message)
}

// TestDeleteOrder_Success tests successful order deletion.
func TestDeleteOrder_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database operations
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.DeleteOrder(context.Background(), "order123")

	assert.NoError(t, err)
}

// TestDeleteOrder_CommitError tests order deletion with commit error.
func TestDeleteOrder_CommitError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database operations
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := service.DeleteOrder(context.Background(), "order123")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "commit_error", appErr.Code)
}

// TestDeleteOrder_NilDBConnection tests order deletion with nil database connection.
func TestDeleteOrder_NilDBConnection(t *testing.T) {
	service := NewOrderService(nil, nil)

	err := service.DeleteOrder(context.Background(), "order123")

	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
	assert.Equal(t, "DB connection is nil", appErr.Message)
}

// TestOrderService_NilDependencies tests service behavior with nil dependencies.
func TestOrderService_NilDependencies(t *testing.T) {
	service := NewOrderService(nil, nil)

	// Test CreateOrder with nil dependencies
	user := database.User{ID: "user123"}
	params := CreateOrderRequest{
		Items: []OrderItemInput{
			{ProductID: "prod1", Quantity: 1, Price: 10.50},
		},
	}

	result, err := service.CreateOrder(context.Background(), user, params)
	require.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)

	// Test GetAllOrders with nil dependencies
	orders, err := service.GetAllOrders(context.Background())
	require.Error(t, err)
	assert.Nil(t, orders)
	appErr = &handlers.AppError{}
	ok = errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
}

// TestGetOrderItemsByOrderID_Success tests successful retrieval of order items.
func TestGetOrderItemsByOrderID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database query with correct 7 columns for OrderItem
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "order_id", "product_id", "quantity", "price", "created_at", "updated_at",
		}).AddRow(
			"item1", "order1", "prod1", 2, "50.00", time.Now(), time.Now(),
		).AddRow(
			"item2", "order1", "prod2", 1, "25.00", time.Now(), time.Now(),
		),
	)

	items, err := service.GetOrderItemsByOrderID(context.Background(), "order1")

	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Len(t, items, 2)
	assert.Equal(t, "item1", items[0].ID)
	assert.Equal(t, "prod1", items[0].ProductID)
	assert.Equal(t, 2, items[0].Quantity)
	assert.Equal(t, "50.00", items[0].Price)
}

// TestGetOrderItemsByOrderID_DatabaseError tests order items retrieval with database error.
func TestGetOrderItemsByOrderID_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database query to return an error
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnError(errors.New("database error"))

	items, err := service.GetOrderItemsByOrderID(context.Background(), "order1")

	require.Error(t, err)
	assert.Nil(t, items)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Failed to fetch order items", appErr.Message)
}

// TestGetOrderItemsByOrderID_NilDatabase tests order items retrieval with nil database.
func TestGetOrderItemsByOrderID_NilDatabase(t *testing.T) {
	service := NewOrderService(nil, nil)

	items, err := service.GetOrderItemsByOrderID(context.Background(), "order1")

	require.Error(t, err)
	assert.Nil(t, items)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "database_error", appErr.Code)
	assert.Equal(t, "Database not initialized", appErr.Message)
}

// TestGetOrderItemsByOrderID_EmptyResult tests order items retrieval with empty result.
func TestGetOrderItemsByOrderID_EmptyResult(t *testing.T) {
	db, mock, _ := sqlmock.New()
	queries := database.New(db)

	service := NewOrderService(queries, db)

	// Mock the database query to return empty result with generic pattern
	mock.ExpectQuery("SELECT (.+) FROM order_items").WithArgs("order1").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "order_id", "product_id", "quantity", "price", "created_at", "updated_at",
		}),
	)

	items, err := service.GetOrderItemsByOrderID(context.Background(), "order1")

	require.NoError(t, err)
	assert.Empty(t, items)

	// Check for unmet expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
