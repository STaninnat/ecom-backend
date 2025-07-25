// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// order_service.go: Provides a complete implementation of the OrderService interface for managing order operations.

// OrderService defines the business logic interface for order operations.
// Provides methods for creating, retrieving, updating, and deleting orders and order items.
type OrderService interface {
	CreateOrder(ctx context.Context, user database.User, params CreateOrderRequest) (*OrderResponse, error)
	GetAllOrders(ctx context.Context) ([]database.Order, error)
	GetUserOrders(ctx context.Context, user database.User) ([]UserOrderResponse, error)
	GetOrderByID(ctx context.Context, orderID string, user database.User) (*OrderDetailResponse, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]OrderItemResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
	DeleteOrder(ctx context.Context, orderID string) error
}

// orderServiceImpl implements OrderService
type orderServiceImpl struct {
	db     *database.Queries
	dbConn *sql.DB
}

// NewOrderService creates a new OrderService instance.
// Accepts a database.Queries and a database connection, and returns an OrderService implementation.
func NewOrderService(db *database.Queries, dbConn *sql.DB) OrderService {
	return &orderServiceImpl{
		db:     db,
		dbConn: dbConn,
	}
}

// CreateOrder handles the business logic for creating a new order.
// Validates the request, calculates totals, creates the order and items, and commits the transaction.
// Returns the created order response or an error.
func (s *orderServiceImpl) CreateOrder(ctx context.Context, user database.User, params CreateOrderRequest) (*OrderResponse, error) {
	if s.dbConn == nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}

	// Validate request
	if len(params.Items) == 0 {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Order must contain at least one item"}
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range params.Items {
		if item.Quantity <= 0 {
			return nil, &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
		}
		if item.Price < 0 {
			return nil, &handlers.AppError{Code: "invalid_request", Message: "Price cannot be negative"}
		}
		totalAmount += float64(item.Quantity) * item.Price
	}

	orderID := utils.NewUUIDString()
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	queries := s.db.WithTx(tx)

	// Create order
	_, err = queries.CreateOrder(ctx, database.CreateOrderParams{
		ID:                orderID,
		UserID:            user.ID,
		TotalAmount:       fmt.Sprintf("%.2f", totalAmount),
		Status:            "pending",
		PaymentMethod:     utils.ToNullString(params.PaymentMethod),
		ExternalPaymentID: utils.ToNullString(params.ExternalPaymentID),
		TrackingNumber:    utils.ToNullString(params.TrackingNumber),
		ShippingAddress:   utils.ToNullString(params.ShippingAddress),
		ContactPhone:      utils.ToNullString(params.ContactPhone),
		CreatedAt:         timeNow,
		UpdatedAt:         timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "create_order_error", Message: "Error creating order", Err: err}
	}

	// Create order items
	for _, item := range params.Items {
		if item.Quantity > math.MaxInt32 || item.Quantity < math.MinInt32 {
			return nil, &handlers.AppError{Code: "quantity_overflow", Message: fmt.Sprintf("Quantity %d exceeds the max limit for int32", item.Quantity)}
		}

		err := queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:        utils.NewUUIDString(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  int32(item.Quantity),
			Price:     fmt.Sprintf("%.2f", item.Price),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
		if err != nil {
			return nil, &handlers.AppError{Code: "create_order_item_error", Message: "Error creating order item", Err: err}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return &OrderResponse{
		Message: "Created order successful",
		OrderID: orderID,
	}, nil
}

// GetAllOrders retrieves all orders (admin only).
// Returns a list of all orders or an error.
func (s *orderServiceImpl) GetAllOrders(ctx context.Context) ([]database.Order, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Database not initialized", Err: errors.New("db is nil")}
	}

	orders, err := s.db.ListAllOrders(ctx)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to list orders", Err: err}
	}

	return orders, nil
}

// GetUserOrders retrieves orders for a specific user.
// Returns a list of user orders with items or an error.
func (s *orderServiceImpl) GetUserOrders(ctx context.Context, user database.User) ([]UserOrderResponse, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Database not initialized", Err: errors.New("db is nil")}
	}

	orders, err := s.db.GetOrderByUserID(ctx, user.ID)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to get orders", Err: err}
	}

	var response []UserOrderResponse
	for _, order := range orders {
		items, err := s.db.GetOrderItemsByOrderID(ctx, order.ID)
		if err != nil {
			return nil, &handlers.AppError{Code: "database_error", Message: "Failed to get order items", Err: err}
		}

		var itemResponses []OrderItemResponse
		for _, item := range items {
			itemResponses = append(itemResponses, OrderItemResponse{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  int(item.Quantity),
				Price:     item.Price,
			})
		}

		response = append(response, UserOrderResponse{
			OrderID:         order.ID,
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			PaymentMethod:   order.PaymentMethod.String,
			TrackingNumber:  order.TrackingNumber.String,
			ShippingAddress: order.ShippingAddress.String,
			ContactPhone:    order.ContactPhone.String,
			CreatedAt:       order.CreatedAt,
			Items:           itemResponses,
		})
	}

	return response, nil
}

// GetOrderByID retrieves a specific order by ID with authorization check.
// Returns detailed order information with items or an error.
func (s *orderServiceImpl) GetOrderByID(ctx context.Context, orderID string, user database.User) (*OrderDetailResponse, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Database not initialized", Err: errors.New("db is nil")}
	}

	order, err := s.db.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "order_not_found", Message: "Order not found", Err: err}
	}

	// Check authorization
	if order.UserID != user.ID && user.Role != "admin" {
		return nil, &handlers.AppError{Code: "unauthorized", Message: "User is not authorized to view this order"}
	}

	items, err := s.db.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to fetch order items", Err: err}
	}

	return &OrderDetailResponse{
		Order: order,
		Items: items,
	}, nil
}

// GetOrderItemsByOrderID retrieves items for a specific order.
// Returns a list of order items or an error.
func (s *orderServiceImpl) GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]OrderItemResponse, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Database not initialized", Err: errors.New("db is nil")}
	}

	items, err := s.db.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Failed to fetch order items", Err: err}
	}

	var response []OrderItemResponse
	for _, item := range items {
		response = append(response, OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
			Price:     item.Price,
		})
	}

	return response, nil
}

// UpdateOrderStatus updates the status of an order.
// Validates the status, updates the order, and commits the transaction. Returns an error if unsuccessful.
func (s *orderServiceImpl) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return &handlers.AppError{Code: "invalid_status", Message: "Invalid order status"}
	}

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	queries := s.db.WithTx(tx)

	err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
		ID:        orderID,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return &handlers.AppError{Code: "update_failed", Message: "Failed to update order status", Err: err}
	}

	err = tx.Commit()
	if err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// DeleteOrder deletes an order by ID.
// Performs the deletion in a transaction and returns an error if unsuccessful.
func (s *orderServiceImpl) DeleteOrder(ctx context.Context, orderID string) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: errors.New("dbConn is nil")}
	}

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	queries := s.db.WithTx(tx)

	err = queries.DeleteOrderByID(ctx, orderID)
	if err != nil {
		return &handlers.AppError{Code: "delete_order_error", Message: "Failed to delete order", Err: err}
	}

	err = tx.Commit()
	if err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// OrderError is an alias for handlers.AppError for order-related errors.
type OrderError = handlers.AppError
