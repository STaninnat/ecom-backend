package carthandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var TTL = 7 * 24 * time.Hour

const (
	GuestCartPrefix = "guest_cart:"
	MaxQuantity     = 1000 // Maximum quantity per item
	MaxCartItems    = 50   // Maximum items in cart
)

// CartMongoAPI defines the interface for MongoDB cart operations
type CartMongoAPI interface {
	AddItemToCart(ctx context.Context, userID string, item models.CartItem) error
	GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	UpdateItemQuantity(ctx context.Context, userID string, productID string, quantity int) error
	RemoveItemFromCart(ctx context.Context, userID string, productID string) error
	ClearCart(ctx context.Context, userID string) error
}

// ProductAPI defines the interface for product operations
type ProductAPI interface {
	GetProductByID(ctx context.Context, productID string) (database.Product, error)
	UpdateProductStock(ctx context.Context, params database.UpdateProductStockParams) error
}

// OrderAPI defines the interface for order operations
type OrderAPI interface {
	CreateOrder(ctx context.Context, params database.CreateOrderParams) error
	CreateOrderItem(ctx context.Context, params database.CreateOrderItemParams) error
}

// DBConnAPI defines the interface for database connection operations
type DBConnAPI interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTxAPI, error)
}

// DBTxAPI defines the interface for database transaction operations
type DBTxAPI interface {
	Commit() error
	Rollback() error
}

// CartRedisAPI defines the interface for Redis cart operations
type CartRedisAPI interface {
	GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error)
	SaveGuestCart(ctx context.Context, sessionID string, cart *models.Cart) error
	UpdateGuestItemQuantity(ctx context.Context, sessionID, productID string, quantity int) error
	RemoveGuestItem(ctx context.Context, sessionID, productID string) error
	DeleteGuestCart(ctx context.Context, sessionID string) error
}

// CartMongoAdapter adapts the CartMongo to CartMongoAPI interface
type CartMongoAdapter struct {
	cartMongo *intmongo.CartMongo
}

// NewCartMongoAdapter creates a new CartMongoAdapter
func NewCartMongoAdapter(cartMongo *intmongo.CartMongo) CartMongoAPI {
	return &CartMongoAdapter{cartMongo: cartMongo}
}

// AddItemToCart adds an item to the user's cart in MongoDB
func (a *CartMongoAdapter) AddItemToCart(ctx context.Context, userID string, item models.CartItem) error {
	return a.cartMongo.AddItemToCart(ctx, userID, item)
}

// GetCartByUserID retrieves a cart by user ID from MongoDB
func (a *CartMongoAdapter) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	return a.cartMongo.GetCartByUserID(ctx, userID)
}

// UpdateItemQuantity updates the quantity of an item in the user's cart in MongoDB
func (a *CartMongoAdapter) UpdateItemQuantity(ctx context.Context, userID string, productID string, quantity int) error {
	return a.cartMongo.UpdateItemQuantity(ctx, userID, productID, quantity)
}

// RemoveItemFromCart removes an item from the user's cart in MongoDB
func (a *CartMongoAdapter) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	return a.cartMongo.RemoveItemFromCart(ctx, userID, productID)
}

// ClearCart clears the user's cart in MongoDB
func (a *CartMongoAdapter) ClearCart(ctx context.Context, userID string) error {
	return a.cartMongo.ClearCart(ctx, userID)
}

// ProductAdapter adapts the database to ProductAPI interface
type ProductAdapter struct {
	db *database.Queries
}

// NewProductAdapter creates a new ProductAdapter
func NewProductAdapter(db *database.Queries) ProductAPI {
	return &ProductAdapter{db: db}
}

// GetProductByID retrieves a product by its ID
func (a *ProductAdapter) GetProductByID(ctx context.Context, productID string) (database.Product, error) {
	return a.db.GetProductByID(ctx, productID)
}

// UpdateProductStock updates the stock for a product
func (a *ProductAdapter) UpdateProductStock(ctx context.Context, params database.UpdateProductStockParams) error {
	return a.db.UpdateProductStock(ctx, params)
}

// OrderAdapter adapts the database to OrderAPI interface
type OrderAdapter struct {
	db *database.Queries
}

// NewOrderAdapter creates a new OrderAdapter
func NewOrderAdapter(db *database.Queries) OrderAPI {
	return &OrderAdapter{db: db}
}

// CreateOrder creates a new order in the database
func (a *OrderAdapter) CreateOrder(ctx context.Context, params database.CreateOrderParams) error {
	_, err := a.db.CreateOrder(ctx, params)
	return err
}

// CreateOrderItem creates a new order item in the database
func (a *OrderAdapter) CreateOrderItem(ctx context.Context, params database.CreateOrderItemParams) error {
	return a.db.CreateOrderItem(ctx, params)
}

// DBConnAdapter adapts the database connection to DBConnAPI interface
type DBConnAdapter struct {
	dbConn *sql.DB
}

// NewDBConnAdapter creates a new DBConnAdapter
func NewDBConnAdapter(dbConn *sql.DB) DBConnAPI {
	return &DBConnAdapter{dbConn: dbConn}
}

// BeginTx begins a new database transaction
func (a *DBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTxAPI, error) {
	tx, err := a.dbConn.BeginTx(ctx, opts)
	return &DBTxAdapter{tx: tx}, err
}

// DBTxAdapter adapts the database transaction to DBTxAPI interface
type DBTxAdapter struct {
	tx *sql.Tx
}

// Commit commits the database transaction
func (a *DBTxAdapter) Commit() error {
	return a.tx.Commit()
}

// Rollback rolls back the database transaction
func (a *DBTxAdapter) Rollback() error {
	return a.tx.Rollback()
}

// cartRedisImpl implements CartRedisAPI
type cartRedisImpl struct {
	redisClient redis.Cmdable
}

// NewCartRedisAPI creates a new CartRedisAPI instance
func NewCartRedisAPI(redisClient redis.Cmdable) CartRedisAPI {
	return &cartRedisImpl{redisClient: redisClient}
}

// GetGuestCart retrieves a guest cart from Redis
func (r *cartRedisImpl) GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	key := GuestCartPrefix + sessionID

	val, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return &models.Cart{Items: []models.CartItem{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get guest cart: %w", err)
	}

	var cart models.Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal guest cart: %w", err)
	}

	return &cart, nil
}

// SaveGuestCart saves a guest cart to Redis
func (r *cartRedisImpl) SaveGuestCart(ctx context.Context, sessionID string, cart *models.Cart) error {
	key := GuestCartPrefix + sessionID

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	err = r.redisClient.Set(ctx, key, data, TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to save guest cart to Redis: %w", err)
	}

	return nil
}

// UpdateGuestItemQuantity updates the quantity of an item in a guest cart in Redis
func (r *cartRedisImpl) UpdateGuestItemQuantity(ctx context.Context, sessionID, productID string, quantity int) error {
	cart, err := r.GetGuestCart(ctx, sessionID)
	if err != nil {
		return err
	}

	updated := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity = quantity
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("item not found")
	}

	return r.SaveGuestCart(ctx, sessionID, cart)
}

// RemoveGuestItem removes an item from a guest cart in Redis
func (r *cartRedisImpl) RemoveGuestItem(ctx context.Context, sessionID, productID string) error {
	cart, err := r.GetGuestCart(ctx, sessionID)
	if err != nil {
		return err
	}

	newItems := make([]models.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}
	cart.Items = newItems

	return r.SaveGuestCart(ctx, sessionID, cart)
}

// DeleteGuestCart deletes a guest cart from Redis
func (r *cartRedisImpl) DeleteGuestCart(ctx context.Context, sessionID string) error {
	key := GuestCartPrefix + sessionID

	return r.redisClient.Del(ctx, key).Err()
}

// cartServiceImpl implements CartService
type cartServiceImpl struct {
	cartMongo CartMongoAPI
	product   ProductAPI
	order     OrderAPI
	dbConn    DBConnAPI
	redis     CartRedisAPI
}

// NewCartService creates a new CartService instance
func NewCartService(
	cartMongo CartMongoAPI,
	product ProductAPI,
	order OrderAPI,
	dbConn DBConnAPI,
	redis CartRedisAPI,
) CartService {
	return &cartServiceImpl{
		cartMongo: cartMongo,
		product:   product,
		order:     order,
		dbConn:    dbConn,
		redis:     redis,
	}
}

// AddItemToUserCart adds an item to a user's cart
func (s *cartServiceImpl) AddItemToUserCart(ctx context.Context, userID string, productID string, quantity int) error {
	// Validate inputs
	if userID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	if quantity <= 0 {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
	}
	if quantity > MaxQuantity {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity exceeds maximum allowed"}
	}

	// Validate product exists and get price
	product, err := s.product.GetProductByID(ctx, productID)
	if err != nil {
		return &handlers.AppError{Code: "product_not_found", Message: "Product not found", Err: err}
	}

	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		return &handlers.AppError{Code: "invalid_price", Message: "Invalid product price format", Err: err}
	}

	item := models.CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
		Name:      product.Name,
	}

	if err := s.cartMongo.AddItemToCart(ctx, userID, item); err != nil {
		return &handlers.AppError{Code: "add_failed", Message: "Failed to add item to cart", Err: err}
	}

	return nil
}

// AddItemToGuestCart adds an item to a guest cart
func (s *cartServiceImpl) AddItemToGuestCart(ctx context.Context, sessionID string, productID string, quantity int) error {
	// Validate inputs
	if sessionID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	if quantity <= 0 {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
	}

	if quantity > MaxQuantity {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity exceeds maximum allowed"}
	}

	// Validate product exists and get price
	product, err := s.product.GetProductByID(ctx, productID)
	if err != nil {
		return &handlers.AppError{Code: "product_not_found", Message: "Product not found", Err: err}
	}

	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		return &handlers.AppError{Code: "invalid_price", Message: "Invalid product price format", Err: err}
	}

	// Get existing cart or create new one
	cart, err := s.redis.GetGuestCart(ctx, sessionID)
	if err != nil {
		return &handlers.AppError{Code: "get_cart_failed", Message: "Failed to get guest cart", Err: err}
	}

	if cart == nil {
		timeNow := time.Now().UTC()
		cart = &models.Cart{
			ID:        sessionID,
			UserID:    "",
			Items:     []models.CartItem{},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}

	// Check if item already exists and update quantity
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity += quantity
			found = true
			break
		}
	}

	if !found {
		if len(cart.Items) >= MaxCartItems {
			return &handlers.AppError{Code: "cart_full", Message: "Cart is full"}
		}
		cart.Items = append(cart.Items, models.CartItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     price,
			Name:      product.Name,
		})
	}

	cart.UpdatedAt = time.Now().UTC()

	if err := s.redis.SaveGuestCart(ctx, sessionID, cart); err != nil {
		return &handlers.AppError{Code: "save_cart_failed", Message: "Failed to save guest cart", Err: err}
	}

	return nil
}

// GetUserCart retrieves a user's cart
func (s *cartServiceImpl) GetUserCart(ctx context.Context, userID string) (*models.Cart, error) {
	if userID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}

	cart, err := s.cartMongo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get user cart", Err: err}
	}
	return cart, nil
}

// GetGuestCart retrieves a guest cart
func (s *cartServiceImpl) GetGuestCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	if sessionID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}

	cart, err := s.redis.GetGuestCart(ctx, sessionID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get guest cart", Err: err}
	}
	return cart, nil
}

// UpdateItemQuantity updates the quantity of an item in a user's cart
func (s *cartServiceImpl) UpdateItemQuantity(ctx context.Context, userID string, productID string, quantity int) error {
	if userID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	if quantity <= 0 {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
	}
	if quantity > MaxQuantity {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity exceeds maximum allowed"}
	}

	if err := s.cartMongo.UpdateItemQuantity(ctx, userID, productID, quantity); err != nil {
		return &handlers.AppError{Code: "update_failed", Message: "Failed to update item quantity", Err: err}
	}
	return nil
}

// UpdateGuestItemQuantity updates the quantity of an item in a guest cart
func (s *cartServiceImpl) UpdateGuestItemQuantity(ctx context.Context, sessionID string, productID string, quantity int) error {
	if sessionID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	if quantity <= 0 {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity must be greater than 0"}
	}
	if quantity > MaxQuantity {
		return &handlers.AppError{Code: "invalid_request", Message: "Quantity exceeds maximum allowed"}
	}

	if err := s.redis.UpdateGuestItemQuantity(ctx, sessionID, productID, quantity); err != nil {
		return &handlers.AppError{Code: "update_failed", Message: "Failed to update guest item quantity", Err: err}
	}
	return nil
}

// RemoveItem removes an item from a user's cart
func (s *cartServiceImpl) RemoveItem(ctx context.Context, userID string, productID string) error {
	if userID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}

	if err := s.cartMongo.RemoveItemFromCart(ctx, userID, productID); err != nil {
		return &handlers.AppError{Code: "remove_failed", Message: "Failed to remove item from cart", Err: err}
	}
	return nil
}

// RemoveGuestItem removes an item from a guest cart
func (s *cartServiceImpl) RemoveGuestItem(ctx context.Context, sessionID string, productID string) error {
	if sessionID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}

	if err := s.redis.RemoveGuestItem(ctx, sessionID, productID); err != nil {
		return &handlers.AppError{Code: "remove_failed", Message: "Failed to remove item from guest cart", Err: err}
	}
	return nil
}

// DeleteUserCart clears a user's cart
func (s *cartServiceImpl) DeleteUserCart(ctx context.Context, userID string) error {
	if userID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}

	if err := s.cartMongo.ClearCart(ctx, userID); err != nil {
		return &handlers.AppError{Code: "clear_failed", Message: "Failed to clear user cart", Err: err}
	}
	return nil
}

// DeleteGuestCart clears a guest cart
func (s *cartServiceImpl) DeleteGuestCart(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}

	if err := s.redis.DeleteGuestCart(ctx, sessionID); err != nil {
		return &handlers.AppError{Code: "clear_failed", Message: "Failed to clear guest cart", Err: err}
	}
	return nil
}

// CheckoutUserCart processes checkout for a user's cart
func (s *cartServiceImpl) CheckoutUserCart(ctx context.Context, userID string) (*CartCheckoutResult, error) {
	if userID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}

	cart, err := s.cartMongo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_cart_failed", Message: "Failed to get user cart", Err: err}
	}

	if cart == nil || len(cart.Items) == 0 {
		return nil, &handlers.AppError{Code: "cart_empty", Message: "Cart is empty"}
	}

	return s.processCheckout(ctx, cart, userID)
}

// CheckoutGuestCart processes checkout for a guest cart
func (s *cartServiceImpl) CheckoutGuestCart(ctx context.Context, sessionID string, userID string) (*CartCheckoutResult, error) {
	if sessionID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "Session ID is required"}
	}
	if userID == "" {
		return nil, &handlers.AppError{Code: "invalid_request", Message: "User ID is required"}
	}

	cart, err := s.redis.GetGuestCart(ctx, sessionID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_cart_failed", Message: "Failed to get guest cart", Err: err}
	}

	if cart == nil || len(cart.Items) == 0 {
		return nil, &handlers.AppError{Code: "cart_empty", Message: "Cart is empty"}
	}

	result, err := s.processCheckout(ctx, cart, userID)
	if err != nil {
		return nil, err
	}

	// Clear guest cart after successful checkout
	if err := s.redis.DeleteGuestCart(ctx, sessionID); err != nil {
		// Log error but don't fail the checkout
		// This could be handled by a background job
	}

	return result, nil
}

// processCheckout handles the common checkout logic
func (s *cartServiceImpl) processCheckout(ctx context.Context, cart *models.Cart, userID string) (*CartCheckoutResult, error) {
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Failed to start transaction", Err: err}
	}
	defer tx.Rollback()

	totalAmount := 0.0
	timeNow := time.Now().UTC()

	// Validate stock and calculate total
	for _, item := range cart.Items {
		qty32, err := safeIntToInt32(item.Quantity)
		if err != nil {
			return nil, &handlers.AppError{Code: "invalid_quantity", Message: "Quantity too large", Err: err}
		}

		product, err := s.product.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, &handlers.AppError{Code: "product_not_found", Message: "Product not found", Err: err}
		}

		if product.Stock < qty32 {
			return nil, &handlers.AppError{Code: "insufficient_stock", Message: "Insufficient stock for product", Err: err}
		}

		totalAmount += item.Price * float64(item.Quantity)
	}

	// Create order
	orderID := uuid.New().String()
	err = s.order.CreateOrder(ctx, database.CreateOrderParams{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: fmt.Sprintf("%.2f", totalAmount),
		Status:      "pending",
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "create_order_failed", Message: "Failed to create order", Err: err}
	}

	// Create order items and update stock
	for _, item := range cart.Items {
		qty32, err := safeIntToInt32(item.Quantity)
		if err != nil {
			return nil, &handlers.AppError{Code: "invalid_quantity", Message: "Quantity too large", Err: err}
		}

		negStock, err := safeIntToInt32(-item.Quantity)
		if err != nil {
			return nil, &handlers.AppError{Code: "invalid_quantity", Message: "Quantity too large", Err: err}
		}

		// Update product stock
		err = s.product.UpdateProductStock(ctx, database.UpdateProductStockParams{
			ID:    item.ProductID,
			Stock: negStock,
		})
		if err != nil {
			return nil, &handlers.AppError{Code: "update_stock_failed", Message: "Failed to update product stock", Err: err}
		}

		// Create order item
		err = s.order.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  qty32,
			Price:     fmt.Sprintf("%.2f", item.Price),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
		if err != nil {
			return nil, &handlers.AppError{Code: "create_order_item_failed", Message: "Failed to create order item", Err: err}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_failed", Message: "Failed to commit transaction", Err: err}
	}

	// Clear user cart
	if err := s.cartMongo.ClearCart(ctx, userID); err != nil {
		// Log error but don't fail the checkout
		// This could be handled by a background job
	}

	return &CartCheckoutResult{
		OrderID: orderID,
		Message: "Order placed successfully",
	}, nil
}

// safeIntToInt32 safely converts int to int32
func safeIntToInt32(i int) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, fmt.Errorf("value %d overflows int32", i)
	}
	return int32(i), nil
}
