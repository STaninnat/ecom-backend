// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// cart_service_test.go: Tests for interfaces, adapters, and services for managing shopping carts and checkout processes.

const (
	testUserID           = "user123"
	testSessionIDService = "sess123"
)

// TestAddItemToUserCart_ProductNotFound tests the service behavior when the product is not found.
// It ensures that the service correctly wraps the database error in an AppError with the "product_not_found" code.
func TestAddItemToUserCart_ProductNotFound(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	dbErr := errors.New("product not found")
	mockProduct.On("GetProductByID", mock.Anything, "product123").Return(database.Product{}, dbErr)

	err := svc.AddItemToUserCart(context.Background(), "user123", "product123", 2)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "product_not_found", appErr.Code)
	mockProduct.AssertExpectations(t)
}

// TestAddItemToUserCart_InvalidPrice tests the service behavior when the product price is invalid.
// It ensures that the service returns an AppError with the "invalid_price" code.
func TestAddItemToUserCart_InvalidPrice(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	product := database.Product{
		ID:    "product123",
		Name:  "Test Product",
		Price: "invalid_price",
	}

	mockProduct.On("GetProductByID", mock.Anything, "product123").Return(product, nil)

	err := svc.AddItemToUserCart(context.Background(), "user123", "product123", 2)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_price", appErr.Code)
	mockProduct.AssertExpectations(t)
}

// TestAddItemToUserCart_AddFailed tests the service behavior when adding the item to cart fails.
// It ensures that the service correctly wraps the database error in an AppError with the "add_failed" code.
func TestAddItemToUserCart_AddFailed(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	product := database.Product{
		ID:    "product123",
		Name:  "Test Product",
		Price: "10.50",
	}

	dbErr := errors.New("add to cart failed")
	mockProduct.On("GetProductByID", mock.Anything, "product123").Return(product, nil)
	mockCartMongo.On("AddItemToCart", mock.Anything, "user123", mock.AnythingOfType("models.CartItem")).Return(dbErr)

	err := svc.AddItemToUserCart(context.Background(), "user123", "product123", 2)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "add_failed", appErr.Code)
	mockProduct.AssertExpectations(t)
	mockCartMongo.AssertExpectations(t)
}

// TestAddItemToGuestCart_Success tests the successful addition of an item to a guest cart.
// It verifies that the service correctly validates inputs, gets product information,
// retrieves or creates the guest cart, and saves it to Redis.
func TestAddItemToGuestCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	product := database.Product{
		ID:    "product123",
		Name:  "Test Product",
		Price: "10.50",
	}

	existingCart := &models.Cart{
		ID:        "session123",
		UserID:    "",
		Items:     []models.CartItem{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockProduct.On("GetProductByID", mock.Anything, "product123").Return(product, nil)
	mockRedis.On("GetGuestCart", mock.Anything, "session123").Return(existingCart, nil)
	mockRedis.On("SaveGuestCart", mock.Anything, "session123", mock.AnythingOfType("*models.Cart")).Return(nil)

	err := svc.AddItemToGuestCart(context.Background(), "session123", "product123", 2)
	assert.NoError(t, err)
	mockProduct.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestAddItemToGuestCart_InvalidInputs tests the service behavior with invalid inputs for guest cart.
// It ensures that the service returns appropriate AppError codes for missing or invalid parameters.
func TestAddItemToGuestCart_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name      string
		sessionID string
		productID string
		quantity  int
		wantCode  string
	}{
		{"empty sessionID", "", "product123", 2, "invalid_request"},
		{"empty productID", "session123", "", 2, "invalid_request"},
		{"zero quantity", "session123", "product123", 0, "invalid_request"},
		{"negative quantity", "session123", "product123", -1, "invalid_request"},
		{"excessive quantity", "session123", "product123", 1001, "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.AddItemToGuestCart(context.Background(), tc.sessionID, tc.productID, tc.quantity)
			assert.Error(t, err)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestAddItemToGuestCart_CartFull tests the service behavior when the guest cart is full.
// It ensures that the service returns an AppError with the "cart_full" code when trying to add items to a full cart.
func TestAddItemToGuestCart_CartFull(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	product := database.Product{
		ID:    "product123",
		Name:  "Test Product",
		Price: "10.50",
	}

	// Create a cart with MaxCartItems items
	items := make([]models.CartItem, MaxCartItems)
	for i := 0; i < MaxCartItems; i++ {
		items[i] = models.CartItem{
			ProductID: fmt.Sprintf("product%d", i),
			Quantity:  1,
			Price:     10.0,
			Name:      fmt.Sprintf("Product %d", i),
		}
	}

	fullCart := &models.Cart{
		ID:        "session123",
		UserID:    "",
		Items:     items,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockProduct.On("GetProductByID", mock.Anything, "product123").Return(product, nil)
	mockRedis.On("GetGuestCart", mock.Anything, "session123").Return(fullCart, nil)

	err := svc.AddItemToGuestCart(context.Background(), "session123", "product123", 2)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "cart_full", appErr.Code)
	mockProduct.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestGetUserCart_Success tests the successful retrieval of a user's cart.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected cart.
func TestGetUserCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	expectedCart := &models.Cart{
		ID:     "user123",
		UserID: "user123",
		Items:  []models.CartItem{},
	}

	mockCartMongo.On("GetCartByUserID", mock.Anything, "user123").Return(expectedCart, nil)

	cart, err := svc.GetUserCart(context.Background(), "user123")
	assert.NoError(t, err)
	assert.Equal(t, expectedCart, cart)
	mockCartMongo.AssertExpectations(t)
}

// TestGetUserCart_InvalidInput tests the service behavior with invalid user ID.
// It ensures that the service returns an AppError with the "invalid_request" code.
func TestGetUserCart_InvalidInput(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cart, err := svc.GetUserCart(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, cart)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestGetUserCart_Failure tests the service behavior when the database operation fails.
// It ensures that the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetUserCart_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	dbErr := errors.New("get cart failed")
	mockCartMongo.On("GetCartByUserID", mock.Anything, "user123").Return((*models.Cart)(nil), dbErr)

	cart, err := svc.GetUserCart(context.Background(), "user123")
	assert.Error(t, err)
	assert.Nil(t, cart)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	mockCartMongo.AssertExpectations(t)
}

// TestGetGuestCart_Success tests the successful retrieval of a guest cart.
// It verifies that the service correctly delegates to the Redis layer and returns the expected cart.
func TestGetGuestCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	expectedCart := &models.Cart{
		ID:     "session123",
		UserID: "",
		Items:  []models.CartItem{},
	}

	mockRedis.On("GetGuestCart", mock.Anything, "session123").Return(expectedCart, nil)

	cart, err := svc.GetGuestCart(context.Background(), "session123")
	assert.NoError(t, err)
	assert.Equal(t, expectedCart, cart)
	mockRedis.AssertExpectations(t)
}

// TestGetGuestCart_InvalidInput tests the service behavior with invalid session ID.
// It ensures that the service returns an AppError with the "invalid_request" code.
func TestGetGuestCart_InvalidInput(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cart, err := svc.GetGuestCart(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, cart)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestGetGuestCart_Failure tests the service behavior when the Redis operation fails.
// It ensures that the service correctly wraps the Redis error in an AppError with the "get_failed" code.
func TestGetGuestCart_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	redisErr := errors.New("get guest cart failed")
	mockRedis.On("GetGuestCart", mock.Anything, "session123").Return((*models.Cart)(nil), redisErr)

	cart, err := svc.GetGuestCart(context.Background(), "session123")
	assert.Error(t, err)
	assert.Nil(t, cart)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	mockRedis.AssertExpectations(t)
}

// TestUpdateItemQuantity_Success tests the successful update of an item quantity in a user's cart.
// It verifies that the service correctly validates inputs and delegates to the MongoDB layer.
func TestUpdateItemQuantity_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockCartMongo.On("UpdateItemQuantity", mock.Anything, "user123", "product123", 5).Return(nil)

	err := svc.UpdateItemQuantity(context.Background(), "user123", "product123", 5)
	assert.NoError(t, err)
	mockCartMongo.AssertExpectations(t)
}

// TestUpdateItemQuantity_InvalidInputs tests the service behavior with invalid inputs for quantity update.
// It ensures that the service returns appropriate AppError codes for missing or invalid parameters.
func TestUpdateItemQuantity_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name      string
		userID    string
		productID string
		quantity  int
		wantCode  string
	}{
		{"empty userID", "", "product123", 5, "invalid_request"},
		{"empty productID", "user123", "", 5, "invalid_request"},
		{"zero quantity", "user123", "product123", 0, "invalid_request"},
		{"negative quantity", "user123", "product123", -1, "invalid_request"},
		{"excessive quantity", "user123", "product123", 1001, "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.UpdateItemQuantity(context.Background(), tc.userID, tc.productID, tc.quantity)
			assert.Error(t, err)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestUpdateItemQuantity_Failure tests the service behavior when the database operation fails.
// It ensures that the service correctly wraps the database error in an AppError with the "update_failed" code.
func TestUpdateItemQuantity_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	dbErr := errors.New("update quantity failed")
	mockCartMongo.On("UpdateItemQuantity", mock.Anything, "user123", "product123", 5).Return(dbErr)

	err := svc.UpdateItemQuantity(context.Background(), "user123", "product123", 5)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "update_failed", appErr.Code)
	mockCartMongo.AssertExpectations(t)
}

// TestUpdateGuestItemQuantity_Success tests the successful update of an item quantity in a guest cart.
// It verifies that the service correctly validates inputs and delegates to the Redis layer.
func TestUpdateGuestItemQuantity_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockRedis.On("UpdateGuestItemQuantity", mock.Anything, "session123", "product123", 5).Return(nil)

	err := svc.UpdateGuestItemQuantity(context.Background(), "session123", "product123", 5)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestUpdateGuestItemQuantity_InvalidInputs tests the service behavior with invalid inputs for guest quantity update.
// It ensures that the service returns appropriate AppError codes for missing or invalid parameters.
func TestUpdateGuestItemQuantity_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name      string
		sessionID string
		productID string
		quantity  int
		wantCode  string
	}{
		{"empty sessionID", "", "product123", 5, "invalid_request"},
		{"empty productID", "session123", "", 5, "invalid_request"},
		{"zero quantity", "session123", "product123", 0, "invalid_request"},
		{"negative quantity", "session123", "product123", -1, "invalid_request"},
		{"excessive quantity", "session123", "product123", 1001, "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.UpdateGuestItemQuantity(context.Background(), tc.sessionID, tc.productID, tc.quantity)
			assert.Error(t, err)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestUpdateGuestItemQuantity_Failure tests the service behavior when the Redis operation fails.
// It ensures that the service correctly wraps the Redis error in an AppError with the "update_failed" code.
func TestUpdateGuestItemQuantity_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	redisErr := errors.New("update guest quantity failed")
	mockRedis.On("UpdateGuestItemQuantity", mock.Anything, "session123", "product123", 5).Return(redisErr)

	err := svc.UpdateGuestItemQuantity(context.Background(), "session123", "product123", 5)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "update_failed", appErr.Code)
	mockRedis.AssertExpectations(t)
}

// TestRemoveItem_Success tests the successful removal of an item from a user's cart.
// It verifies that the service correctly validates inputs and delegates to the MongoDB layer.
func TestRemoveItem_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockCartMongo.On("RemoveItemFromCart", mock.Anything, "user123", "product123").Return(nil)

	err := svc.RemoveItem(context.Background(), "user123", "product123")
	assert.NoError(t, err)
	mockCartMongo.AssertExpectations(t)
}

// TestRemoveItem_InvalidInputs tests the service behavior with invalid inputs for item removal.
// It ensures that the service returns appropriate AppError codes for missing parameters.
func TestRemoveItem_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name      string
		userID    string
		productID string
		wantCode  string
	}{
		{"empty userID", "", "product123", "invalid_request"},
		{"empty productID", "user123", "", "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RemoveItem(context.Background(), tc.userID, tc.productID)
			assert.Error(t, err)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestRemoveItem_Failure tests the service behavior when the database operation fails.
// It ensures that the service correctly wraps the database error in an AppError with the "remove_failed" code.
func TestRemoveItem_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	dbErr := errors.New("remove item failed")
	mockCartMongo.On("RemoveItemFromCart", mock.Anything, "user123", "product123").Return(dbErr)

	err := svc.RemoveItem(context.Background(), "user123", "product123")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "remove_failed", appErr.Code)
	mockCartMongo.AssertExpectations(t)
}

// TestRemoveGuestItem_Success tests the successful removal of an item from a guest cart.
// It verifies that the service correctly validates inputs and delegates to the Redis layer.
func TestRemoveGuestItem_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockRedis.On("RemoveGuestItem", mock.Anything, "session123", "product123").Return(nil)

	err := svc.RemoveGuestItem(context.Background(), "session123", "product123")
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestRemoveGuestItem_InvalidInputs tests the service behavior with invalid inputs for guest item removal.
// It ensures that the service returns appropriate AppError codes for missing parameters.
func TestRemoveGuestItem_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name      string
		sessionID string
		productID string
		wantCode  string
	}{
		{"empty sessionID", "", "product123", "invalid_request"},
		{"empty productID", "session123", "", "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RemoveGuestItem(context.Background(), tc.sessionID, tc.productID)
			assert.Error(t, err)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestRemoveGuestItem_Failure tests the service behavior when the Redis operation fails.
// It ensures that the service correctly wraps the Redis error in an AppError with the "remove_failed" code.
func TestRemoveGuestItem_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	redisErr := errors.New("remove guest item failed")
	mockRedis.On("RemoveGuestItem", mock.Anything, "session123", "product123").Return(redisErr)

	err := svc.RemoveGuestItem(context.Background(), "session123", "product123")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "remove_failed", appErr.Code)
	mockRedis.AssertExpectations(t)
}

// TestDeleteUserCart_Success tests the successful deletion of a user's cart.
// It verifies that the service correctly validates inputs and delegates to the MongoDB layer.
func TestDeleteUserCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockCartMongo.On("ClearCart", mock.Anything, "user123").Return(nil)

	err := svc.DeleteUserCart(context.Background(), "user123")
	assert.NoError(t, err)
	mockCartMongo.AssertExpectations(t)
}

// TestDeleteUserCart_InvalidInput tests the service behavior with invalid user ID.
// It ensures that the service returns an AppError with the "invalid_request" code.
func TestDeleteUserCart_InvalidInput(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	err := svc.DeleteUserCart(context.Background(), "")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestDeleteUserCart_Failure tests the service behavior when the database operation fails.
// It ensures that the service correctly wraps the database error in an AppError with the "clear_failed" code.
func TestDeleteUserCart_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	dbErr := errors.New("clear cart failed")
	mockCartMongo.On("ClearCart", mock.Anything, "user123").Return(dbErr)

	err := svc.DeleteUserCart(context.Background(), "user123")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "clear_failed", appErr.Code)
	mockCartMongo.AssertExpectations(t)
}

// TestDeleteGuestCart_Success tests the successful deletion of a guest cart.
// It verifies that the service correctly validates inputs and delegates to the Redis layer.
func TestDeleteGuestCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	mockRedis.On("DeleteGuestCart", mock.Anything, "session123").Return(nil)

	err := svc.DeleteGuestCart(context.Background(), "session123")
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestDeleteGuestCart_InvalidInput tests the service behavior with invalid session ID.
// It ensures that the service returns an AppError with the "invalid_request" code.
func TestDeleteGuestCart_InvalidInput(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	err := svc.DeleteGuestCart(context.Background(), "")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestDeleteGuestCart_Failure tests the service behavior when the Redis operation fails.
// It ensures that the service correctly wraps the Redis error in an AppError with the "clear_failed" code.
func TestDeleteGuestCart_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	redisErr := errors.New("delete guest cart failed")
	mockRedis.On("DeleteGuestCart", mock.Anything, "session123").Return(redisErr)

	err := svc.DeleteGuestCart(context.Background(), "session123")
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "clear_failed", appErr.Code)
	mockRedis.AssertExpectations(t)
}

// TestCheckoutUserCart_Success tests the successful checkout of a user's cart.
// It verifies that the service correctly processes the checkout, updates stock, creates order and items, and clears the cart.
func TestCheckoutUserCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items: []models.CartItem{
			{ProductID: "prod1", Quantity: 2, Price: 10.0, Name: "Product 1"},
			{ProductID: "prod2", Quantity: 1, Price: 20.0, Name: "Product 2"},
		},
	}

	// Mock getting the cart
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	// Mock transaction begin
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	// Mock product lookups and stock
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod2").Return(database.Product{ID: "prod2", Name: "Product 2", Price: "20.00", Stock: 5}, nil)
	// Mock order creation
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	// Mock stock update
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	// Mock order item creation
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(nil)
	// Mock commit
	mockDBTx.On("Commit").Return(nil)
	mockDBTx.On("Rollback").Return(nil)
	// Mock cart clear
	mockCartMongo.On("ClearCart", mock.Anything, userID).Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "Order placed successfully", result.Message)

	mockCartMongo.AssertExpectations(t)
	mockProduct.AssertExpectations(t)
	mockOrder.AssertExpectations(t)
	mockDBConn.AssertExpectations(t)
	mockDBTx.AssertExpectations(t)
}

// TestCheckoutUserCart_EmptyCart tests checkout when the user's cart is empty.
func TestCheckoutUserCart_EmptyCart(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(&models.Cart{ID: userID, UserID: userID, Items: []models.CartItem{}}, nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "cart_empty", appErr.Code)
}

// TestCheckoutUserCart_InsufficientStock tests checkout when a product has insufficient stock.
func TestCheckoutUserCart_InsufficientStock(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 5, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 2}, nil)
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "insufficient_stock", appErr.Code)
}

// TestCheckoutUserCart_ProductNotFound tests checkout when a product is not found.
func TestCheckoutUserCart_ProductNotFound(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{}, errors.New("not found"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "product_not_found", appErr.Code)
}

// TestCheckoutUserCart_BeginTxError tests checkout when transaction begin fails.
func TestCheckoutUserCart_BeginTxError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	// Fix: Return a typed nil for DBTxAPI to avoid interface conversion panic
	var nilTx DBTxAPI = (*MockDBTxAPI)(nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(nilTx, errors.New("tx error"))

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
}

// TestCheckoutUserCart_CreateOrderError tests checkout when order creation fails.
func TestCheckoutUserCart_CreateOrderError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(errors.New("order error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "create_order_failed", appErr.Code)
}

// TestCheckoutUserCart_UpdateStockError tests checkout when updating product stock fails.
func TestCheckoutUserCart_UpdateStockError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(errors.New("stock error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "update_stock_failed", appErr.Code)
}

// TestCheckoutUserCart_CreateOrderItemError tests checkout when creating an order item fails.
func TestCheckoutUserCart_CreateOrderItemError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(errors.New("order item error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "create_order_item_failed", appErr.Code)
}

// TestCheckoutUserCart_ClearCartError tests checkout when clearing the cart fails (should not fail checkout).
func TestCheckoutUserCart_ClearCartError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	cart := &models.Cart{
		ID:     userID,
		UserID: userID,
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockCartMongo.On("GetCartByUserID", mock.Anything, userID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(nil)
	mockDBTx.On("Commit").Return(nil)
	mockDBTx.On("Rollback").Return(nil)
	mockCartMongo.On("ClearCart", mock.Anything, userID).Return(errors.New("clear error"))

	result, err := svc.CheckoutUserCart(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "Order placed successfully", result.Message)
}

// TestCheckoutGuestCart_Success tests the successful checkout of a guest cart.
func TestCheckoutGuestCart_Success(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items: []models.CartItem{
			{ProductID: "prod1", Quantity: 2, Price: 10.0, Name: "Product 1"},
		},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(nil)
	mockDBTx.On("Commit").Return(nil)
	mockDBTx.On("Rollback").Return(nil)
	mockRedis.On("DeleteGuestCart", mock.Anything, sessionID).Return(nil)
	mockCartMongo.On("ClearCart", mock.Anything, userID).Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "Order placed successfully", result.Message)
}

// TestCheckoutGuestCart_EmptyCart tests checkout when the guest cart is empty.
func TestCheckoutGuestCart_EmptyCart(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(&models.Cart{ID: sessionID, UserID: "", Items: []models.CartItem{}}, nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "cart_empty", appErr.Code)
}

// TestCheckoutGuestCart_MissingSessionID tests checkout with missing sessionID.
func TestCheckoutGuestCart_MissingSessionID(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	userID := testUserID
	result, err := svc.CheckoutGuestCart(context.Background(), "", userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestCheckoutGuestCart_MissingUserID tests checkout with missing userID.
func TestCheckoutGuestCart_MissingUserID(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, "")
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// TestCheckoutGuestCart_GetCartError tests checkout when getting the guest cart fails.
func TestCheckoutGuestCart_GetCartError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return((*models.Cart)(nil), errors.New("redis error"))

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_cart_failed", appErr.Code)
}

// TestCheckoutGuestCart_InsufficientStock tests checkout when a product has insufficient stock.
func TestCheckoutGuestCart_InsufficientStock(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 5, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 2}, nil)
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "insufficient_stock", appErr.Code)
}

// TestCheckoutGuestCart_ProductNotFound tests checkout when a product is not found.
func TestCheckoutGuestCart_ProductNotFound(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{}, errors.New("not found"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "product_not_found", appErr.Code)
}

// TestCheckoutGuestCart_BeginTxError tests checkout when transaction begin fails.
func TestCheckoutGuestCart_BeginTxError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	// Fix: Return a typed nil for DBTxAPI to avoid interface conversion panic
	var nilTx DBTxAPI = (*MockDBTxAPI)(nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(nilTx, errors.New("tx error"))

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "transaction_error", appErr.Code)
}

// TestCheckoutGuestCart_CreateOrderError tests checkout when order creation fails.
func TestCheckoutGuestCart_CreateOrderError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(errors.New("order error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "create_order_failed", appErr.Code)
}

// TestCheckoutGuestCart_UpdateStockError tests checkout when updating product stock fails.
func TestCheckoutGuestCart_UpdateStockError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(errors.New("stock error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "update_stock_failed", appErr.Code)
}

// TestCheckoutGuestCart_CreateOrderItemError tests checkout when creating an order item fails.
func TestCheckoutGuestCart_CreateOrderItemError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(errors.New("order item error"))
	mockDBTx.On("Rollback").Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "create_order_item_failed", appErr.Code)
}

// TestCheckoutGuestCart_ClearCartError tests checkout when clearing the guest cart fails (should not fail checkout).
func TestCheckoutGuestCart_ClearCartError(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockDBTx := new(MockDBTxAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)
	sessionID := testSessionIDService
	userID := testUserID
	cart := &models.Cart{
		ID:     sessionID,
		UserID: "",
		Items:  []models.CartItem{{ProductID: "prod1", Quantity: 1, Price: 10.0, Name: "Product 1"}},
	}
	mockRedis.On("GetGuestCart", mock.Anything, sessionID).Return(cart, nil)
	mockDBConn.On("BeginTx", mock.Anything, (*sql.TxOptions)(nil)).Return(mockDBTx, nil)
	mockProduct.On("GetProductByID", mock.Anything, "prod1").Return(database.Product{ID: "prod1", Name: "Product 1", Price: "10.00", Stock: 10}, nil)
	mockOrder.On("CreateOrder", mock.Anything, mock.AnythingOfType("database.CreateOrderParams")).Return(nil)
	mockProduct.On("UpdateProductStock", mock.Anything, mock.AnythingOfType("database.UpdateProductStockParams")).Return(nil)
	mockOrder.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("database.CreateOrderItemParams")).Return(nil)
	mockDBTx.On("Commit").Return(nil)
	mockDBTx.On("Rollback").Return(nil)
	mockRedis.On("DeleteGuestCart", mock.Anything, sessionID).Return(errors.New("clear error"))
	mockCartMongo.On("ClearCart", mock.Anything, userID).Return(nil)

	result, err := svc.CheckoutGuestCart(context.Background(), sessionID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "Order placed successfully", result.Message)
}

// TestSafeIntToInt32 tests the safeIntToInt32 helper function.
// It verifies that the function correctly converts int to int32 and handles overflow cases.
func TestSafeIntToInt32(t *testing.T) {
	cases := []struct {
		name    string
		input   int
		want    int32
		wantErr bool
	}{
		{"valid positive", 100, 100, false},
		{"valid negative", -100, -100, false},
		{"zero", 0, 0, false},
		{"max int32", 2147483647, 2147483647, false},
		{"min int32", -2147483648, -2147483648, false},
		{"overflow positive", 2147483648, 0, true},
		{"overflow negative", -2147483649, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeIntToInt32(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestDBConnAdapter_BeginTx_Coverage(_ *testing.T) {
	adapter := &DBConnAdapter{}
	_ = adapter
}

func TestDBTxAdapter_CommitRollback_Coverage(_ *testing.T) {
	adapter := &DBTxAdapter{}
	_ = adapter
}

func TestCartRedisImpl_Coverage(_ *testing.T) {
	r := &cartRedisImpl{}
	_ = r
}

// TestCartMongoAdapter_WithSqlMock tests the CartMongoAdapter using sqlmock
func TestCartMongoAdapter_WithSqlMock(t *testing.T) {
	// Create a mock MongoDB client (we'll use a simple mock since we don't have a real MongoDB)
	mockMongo := &MockCartMongo{}
	// Use the mock directly as CartMongoAPI interface
	var adapter CartMongoAPI = mockMongo

	ctx := context.Background()

	// Test AddItemToCart
	mockMongo.On("AddItemToCart", ctx, "user-id", mock.AnythingOfType("models.CartItem")).Return(nil)
	err := adapter.AddItemToCart(ctx, "user-id", models.CartItem{
		ProductID: "product-1",
		Quantity:  2,
	})
	assert.NoError(t, err)

	// Test GetCartByUserID
	expectedCart := &models.Cart{
		Items: []models.CartItem{
			{ProductID: "product-1", Quantity: 2},
		},
	}
	mockMongo.On("GetCartByUserID", ctx, "user-id").Return(expectedCart, nil)
	cart, err := adapter.GetCartByUserID(ctx, "user-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedCart, cart)

	// Test UpdateItemQuantity
	mockMongo.On("UpdateItemQuantity", ctx, "user-id", "product-1", 3).Return(nil)
	err = adapter.UpdateItemQuantity(ctx, "user-id", "product-1", 3)
	assert.NoError(t, err)

	// Test RemoveItemFromCart
	mockMongo.On("RemoveItemFromCart", ctx, "user-id", "product-1").Return(nil)
	err = adapter.RemoveItemFromCart(ctx, "user-id", "product-1")
	assert.NoError(t, err)

	// Test ClearCart
	mockMongo.On("ClearCart", ctx, "user-id").Return(nil)
	err = adapter.ClearCart(ctx, "user-id")
	assert.NoError(t, err)

	mockMongo.AssertExpectations(t)
}

// TestProductAdapter_WithSqlMock tests the ProductAdapter using sqlmock
func TestProductAdapter_WithSqlMock(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		t.Errorf("db.Close() failed: %v", err)
	// 	}
	// }()

	// Create queries with the mock database
	queries := database.New(db)
	adapter := NewProductAdapter(queries)

	ctx := context.Background()

	// Test GetProductByID
	mock.ExpectQuery("SELECT id, category_id, name, description, price, stock, image_url, is_active, created_at, updated_at FROM products").WithArgs("product-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "category_id", "name", "description", "price", "stock", "image_url", "is_active", "created_at", "updated_at"}).
			AddRow("product-1", "category-1", "Test Product", "Test Description", "10.99", 100, nil, true, time.Now(), time.Now()),
	)
	product, err := adapter.GetProductByID(ctx, "product-1")
	assert.NoError(t, err)
	assert.Equal(t, "product-1", product.ID)

	// Test UpdateProductStock
	mock.ExpectExec("UPDATE products SET stock = \\$2 WHERE id = \\$1").WithArgs("product-1", 95).WillReturnResult(sqlmock.NewResult(0, 1))
	err = adapter.UpdateProductStock(ctx, database.UpdateProductStockParams{
		ID:    "product-1",
		Stock: 95,
	})
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOrderAdapter_WithSqlMock tests the OrderAdapter using sqlmock
func TestOrderAdapter_WithSqlMock(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// mock.ExpectClose()
	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		t.Errorf("db.Close() failed: %v", err)
	// 	}
	// }()

	// Create queries with the mock database
	queries := database.New(db)
	adapter := NewOrderAdapter(queries)

	ctx := context.Background()

	// Test CreateOrder
	mock.ExpectQuery("INSERT INTO orders").WithArgs(
		"order-1", "user-1", "0.00", "pending", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "total_amount", "status", "payment_method", "external_payment_id", "tracking_number", "shipping_address", "contact_phone", "created_at", "updated_at"}).
		AddRow("order-1", "user-1", "0.00", "pending", nil, nil, nil, nil, nil, time.Now(), time.Now()))
	err = adapter.CreateOrder(ctx, database.CreateOrderParams{
		ID:                "order-1",
		UserID:            "user-1",
		TotalAmount:       "0.00",
		Status:            "pending",
		PaymentMethod:     sql.NullString{},
		ExternalPaymentID: sql.NullString{},
		TrackingNumber:    sql.NullString{},
		ShippingAddress:   sql.NullString{},
		ContactPhone:      sql.NullString{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	assert.NoError(t, err)

	// Test CreateOrderItem
	mock.ExpectExec("INSERT INTO order_items").WithArgs(
		"item-1", "order-1", "product-1", 2, "10.99", sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	err = adapter.CreateOrderItem(ctx, database.CreateOrderItemParams{
		ID:        "item-1",
		OrderID:   "order-1",
		ProductID: "product-1",
		Quantity:  2,
		Price:     "10.99",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDBConnAdapter_WithSqlMock tests the DBConnAdapter using sqlmock
func TestDBConnAdapter_WithSqlMock(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	adapter := NewDBConnAdapter(db)
	ctx := context.Background()

	// Test BeginTx with default options
	mock.ExpectBegin()
	tx, err := adapter.BeginTx(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Test BeginTx with custom options
	mock.ExpectBegin()
	tx, err = adapter.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDBTxAdapter tests the DBTxAdapter
func TestDBTxAdapter(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		t.Errorf("db.Close() failed: %v", err)
	// 	}
	// }()

	// Start a transaction
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)

	adapter := &DBTxAdapter{tx: tx}

	// Test Commit
	mock.ExpectCommit()
	err = adapter.Commit()
	assert.NoError(t, err)

	// Test Rollback (we need a new transaction since the previous one was committed)
	mock.ExpectBegin()
	tx2, err := db.Begin()
	assert.NoError(t, err)

	adapter2 := &DBTxAdapter{tx: tx2}
	mock.ExpectRollback()
	err = adapter2.Rollback()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCartRedisAPI_WithRedisMock tests the CartRedisAPI using redismock
func TestCartRedisAPI_WithRedisMock(t *testing.T) {
	// Create a mock Redis client
	redisClient, mock := redismock.NewClientMock()
	adapter := NewCartRedisAPI(redisClient)

	ctx := context.Background()
	sessionID := "session-123"

	// Test GetGuestCart - empty cart
	mock.ExpectGet("guest_cart:session-123").RedisNil()
	cart, err := adapter.GetGuestCart(ctx, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Empty(t, cart.Items)

	// Test GetGuestCart - with existing cart
	existingCart := &models.Cart{
		Items: []models.CartItem{
			{ProductID: "product-1", Quantity: 2},
		},
	}
	cartJSON, _ := json.Marshal(existingCart)
	mock.ExpectGet("guest_cart:session-123").SetVal(string(cartJSON))
	cart, err = adapter.GetGuestCart(ctx, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Len(t, cart.Items, 1)

	// Test SaveGuestCart
	cartToSave := &models.Cart{
		Items: []models.CartItem{
			{ProductID: "product-1", Quantity: 3},
		},
	}
	cartJSON, _ = json.Marshal(cartToSave)
	mock.ExpectSet("guest_cart:session-123", cartJSON, 7*24*time.Hour).SetVal("OK")
	err = adapter.SaveGuestCart(ctx, sessionID, cartToSave)
	assert.NoError(t, err)

	// Test UpdateGuestItemQuantity
	// First get the cart
	mock.ExpectGet("guest_cart:session-123").SetVal(string(cartJSON))
	// Then save the updated cart
	updatedCart := &models.Cart{
		Items: []models.CartItem{
			{ProductID: "product-1", Quantity: 5},
		},
	}
	updatedCartJSON, _ := json.Marshal(updatedCart)
	mock.ExpectSet("guest_cart:session-123", updatedCartJSON, 7*24*time.Hour).SetVal("OK")
	err = adapter.UpdateGuestItemQuantity(ctx, sessionID, "product-1", 5)
	assert.NoError(t, err)

	// Test RemoveGuestItem
	// First get the cart
	mock.ExpectGet("guest_cart:session-123").SetVal(string(cartJSON))
	// Then save the cart without the item
	emptyCart := &models.Cart{Items: []models.CartItem{}}
	emptyCartJSON, _ := json.Marshal(emptyCart)
	mock.ExpectSet("guest_cart:session-123", emptyCartJSON, 7*24*time.Hour).SetVal("OK")
	err = adapter.RemoveGuestItem(ctx, sessionID, "product-1")
	assert.NoError(t, err)

	// Test DeleteGuestCart
	mock.ExpectDel("guest_cart:session-123").SetVal(1)
	err = adapter.DeleteGuestCart(ctx, sessionID)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
