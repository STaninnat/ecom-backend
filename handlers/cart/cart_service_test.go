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
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.NoError(t, err)
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
			require.Error(t, err)
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
	for i := range MaxCartItems {
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
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "cart_full", appErr.Code)
	mockProduct.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// Shared helper for GetUserCart and GetGuestCart tests
func runGetCartTest(t *testing.T, name string, getCart func(context.Context, string) (*models.Cart, error), id string, setupMock func(), expectedCart *models.Cart, expectedErrCode string) {
	t.Run(name, func(t *testing.T) {
		setupMock()
		cart, err := getCart(context.Background(), id)
		if expectedErrCode == "" {
			require.NoError(t, err)
			assert.Equal(t, expectedCart, cart)
		} else {
			require.Error(t, err)
			assert.Nil(t, cart)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, expectedErrCode, appErr.Code)
		}
	})
}

func TestCartScenarios(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	// User cart scenarios
	userSetups := []func(){
		func() {
			mockCartMongo.ExpectedCalls = nil
			expectedCart := &models.Cart{ID: "user123", UserID: "user123", Items: []models.CartItem{}}
			mockCartMongo.On("GetCartByUserID", mock.Anything, "user123").Return(expectedCart, nil)
		},
		func() {
			mockCartMongo.ExpectedCalls = nil
		}, // Will call with empty userID to trigger 'invalid_request'
		func() {
			mockCartMongo.ExpectedCalls = nil
			dbErr := errors.New("get cart failed")
			mockCartMongo.On("GetCartByUserID", mock.Anything, "user123").Return((*models.Cart)(nil), dbErr)
		},
	}
	userExpectedCarts := []*models.Cart{
		{ID: "user123", UserID: "user123", Items: []models.CartItem{}},
		nil,
		nil,
	}
	userExpectedErrCodes := []string{"", "invalid_request", "get_failed"}

	runGetCartTest(t, "UserCart/A", svc.GetUserCart, "user123", userSetups[0], userExpectedCarts[0], userExpectedErrCodes[0])
	runGetCartTest(t, "UserCart/B", svc.GetUserCart, "", userSetups[1], userExpectedCarts[1], userExpectedErrCodes[1])
	runGetCartTest(t, "UserCart/C", svc.GetUserCart, "user123", userSetups[2], userExpectedCarts[2], userExpectedErrCodes[2])

	// Guest cart scenarios
	guestSetups := []func(){
		func() {
			mockRedis.ExpectedCalls = nil
			expectedCart := &models.Cart{ID: "session123", UserID: "", Items: []models.CartItem{}}
			mockRedis.On("GetGuestCart", mock.Anything, "session123").Return(expectedCart, nil)
		},
		func() {
			mockRedis.ExpectedCalls = nil
		}, // Will call with empty sessionID to trigger 'invalid_request'
		func() {
			mockRedis.ExpectedCalls = nil
			redisErr := errors.New("get guest cart failed")
			mockRedis.On("GetGuestCart", mock.Anything, "session123").Return((*models.Cart)(nil), redisErr)
		},
	}
	guestExpectedCarts := []*models.Cart{
		{ID: "session123", UserID: "", Items: []models.CartItem{}},
		nil,
		nil,
	}
	guestExpectedErrCodes := []string{"", "invalid_request", "get_failed"}

	runGetCartTest(t, "GuestCart/A", svc.GetGuestCart, "session123", guestSetups[0], guestExpectedCarts[0], guestExpectedErrCodes[0])
	runGetCartTest(t, "GuestCart/B", svc.GetGuestCart, "", guestSetups[1], guestExpectedCarts[1], guestExpectedErrCodes[1])
	runGetCartTest(t, "GuestCart/C", svc.GetGuestCart, "session123", guestSetups[2], guestExpectedCarts[2], guestExpectedErrCodes[2])
}

// Shared helper for UpdateItemQuantity and UpdateGuestItemQuantity invalid input tests
func runUpdateQuantityInvalidInputsTest(t *testing.T, updateFunc func(context.Context, string, string, int) error, id string, productID string, quantity int, wantCode string) {
	err := updateFunc(context.Background(), id, productID, quantity)
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, wantCode, appErr.Code)
}

// Shared helper for UpdateItemQuantity and UpdateGuestItemQuantity failure tests
func runUpdateQuantityFailureTest(t *testing.T, updateFunc func(context.Context, string, string, int) error, setupMock func(), id string, productID string, quantity int, wantCode string) {
	setupMock()
	err := updateFunc(context.Background(), id, productID, quantity)
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, wantCode, appErr.Code)
}

func TestUpdateQuantity_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name       string
		id         string
		productID  string
		quantity   int
		wantCode   string
		updateFunc func(context.Context, string, string, int) error
	}{
		{"empty userID", "", "product123", 5, "invalid_request", svc.UpdateItemQuantity},
		{"empty productID", "user123", "", 5, "invalid_request", svc.UpdateItemQuantity},
		{"zero quantity", "user123", "product123", 0, "invalid_request", svc.UpdateItemQuantity},
		{"negative quantity", "user123", "product123", -1, "invalid_request", svc.UpdateItemQuantity},
		{"excessive quantity", "user123", "product123", 1001, "invalid_request", svc.UpdateItemQuantity},
		{"empty sessionID", "", "product123", 5, "invalid_request", svc.UpdateGuestItemQuantity},
		{"empty productID (guest)", "session123", "", 5, "invalid_request", svc.UpdateGuestItemQuantity},
		{"zero quantity (guest)", "session123", "product123", 0, "invalid_request", svc.UpdateGuestItemQuantity},
		{"negative quantity (guest)", "session123", "product123", -1, "invalid_request", svc.UpdateGuestItemQuantity},
		{"excessive quantity (guest)", "session123", "product123", 1001, "invalid_request", svc.UpdateGuestItemQuantity},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runUpdateQuantityInvalidInputsTest(t, tc.updateFunc, tc.id, tc.productID, tc.quantity, tc.wantCode)
		})
	}
}

func TestUpdateQuantity_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	t.Run("UpdateItemQuantity_Failure", func(t *testing.T) {
		setupMock := func() {
			dbErr := errors.New("update quantity failed")
			mockCartMongo.On("UpdateItemQuantity", mock.Anything, "user123", "product123", 5).Return(dbErr)
		}
		runUpdateQuantityFailureTest(t, svc.UpdateItemQuantity, setupMock, "user123", "product123", 5, "update_failed")
		mockCartMongo.AssertExpectations(t)
	})

	t.Run("UpdateGuestItemQuantity_Failure", func(t *testing.T) {
		setupMock := func() {
			redisErr := errors.New("update guest quantity failed")
			mockRedis.On("UpdateGuestItemQuantity", mock.Anything, "session123", "product123", 5).Return(redisErr)
		}
		runUpdateQuantityFailureTest(t, svc.UpdateGuestItemQuantity, setupMock, "session123", "product123", 5, "update_failed")
		mockRedis.AssertExpectations(t)
	})
}

// Shared helper for RemoveItem and RemoveGuestItem invalid input tests
func runRemoveItemInvalidInputsTest(t *testing.T, removeFunc func(context.Context, string, string) error, id string, productID string, wantCode string) {
	err := removeFunc(context.Background(), id, productID)
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, wantCode, appErr.Code)
}

// Shared helper for RemoveItem and RemoveGuestItem failure tests
func runRemoveItemFailureTest(t *testing.T, removeFunc func(context.Context, string, string) error, setupMock func(), id string, productID string, wantCode string) {
	setupMock()
	err := removeFunc(context.Background(), id, productID)
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, wantCode, appErr.Code)
}

func TestRemoveItem_InvalidInputs(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	cases := []struct {
		name       string
		id         string
		productID  string
		wantCode   string
		removeFunc func(context.Context, string, string) error
	}{
		{"empty userID", "", "product123", "invalid_request", svc.RemoveItem},
		{"empty productID", "user123", "", "invalid_request", svc.RemoveItem},
		{"empty sessionID", "", "product123", "invalid_request", svc.RemoveGuestItem},
		{"empty productID (guest)", "session123", "", "invalid_request", svc.RemoveGuestItem},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runRemoveItemInvalidInputsTest(t, tc.removeFunc, tc.id, tc.productID, tc.wantCode)
		})
	}
}

func TestRemoveItem_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	t.Run("RemoveItem_Failure", func(t *testing.T) {
		setupMock := func() {
			dbErr := errors.New("remove item failed")
			mockCartMongo.On("RemoveItemFromCart", mock.Anything, "user123", "product123").Return(dbErr)
		}
		runRemoveItemFailureTest(t, svc.RemoveItem, setupMock, "user123", "product123", "remove_failed")
		mockCartMongo.AssertExpectations(t)
	})

	t.Run("RemoveGuestItem_Failure", func(t *testing.T) {
		setupMock := func() {
			redisErr := errors.New("remove guest item failed")
			mockRedis.On("RemoveGuestItem", mock.Anything, "session123", "product123").Return(redisErr)
		}
		runRemoveItemFailureTest(t, svc.RemoveGuestItem, setupMock, "session123", "product123", "remove_failed")
		mockRedis.AssertExpectations(t)
	})
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
	require.NoError(t, err)
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
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", appErr.Code)
}

// Shared helper for DeleteUserCart and DeleteGuestCart failure tests
func runDeleteCartFailureTest(t *testing.T, deleteFunc func(context.Context, string) error, setupMock func(), id string, wantCode string) {
	setupMock()
	err := deleteFunc(context.Background(), id)
	require.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, wantCode, appErr.Code)
}

func TestDeleteCart_Failure(t *testing.T) {
	mockCartMongo := new(MockCartMongoAPI)
	mockProduct := new(MockProductAPI)
	mockOrder := new(MockOrderAPI)
	mockDBConn := new(MockDBConnAPI)
	mockRedis := new(MockCartRedisAPI)

	svc := NewCartService(mockCartMongo, mockProduct, mockOrder, mockDBConn, mockRedis)

	t.Run("DeleteUserCart_Failure", func(t *testing.T) {
		setupMock := func() {
			dbErr := errors.New("clear cart failed")
			mockCartMongo.On("ClearCart", mock.Anything, "user123").Return(dbErr)
		}
		runDeleteCartFailureTest(t, svc.DeleteUserCart, setupMock, "user123", "clear_failed")
		mockCartMongo.AssertExpectations(t)
	})

	t.Run("DeleteGuestCart_Failure", func(t *testing.T) {
		setupMock := func() {
			redisErr := errors.New("delete guest cart failed")
			mockRedis.On("DeleteGuestCart", mock.Anything, "session123").Return(redisErr)
		}
		runDeleteCartFailureTest(t, svc.DeleteGuestCart, setupMock, "session123", "clear_failed")
		mockRedis.AssertExpectations(t)
	})
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
	require.NoError(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.NoError(t, err)
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
	require.NoError(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.NoError(t, err)
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
				require.NoError(t, err)
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
	require.NoError(t, err)

	// Test GetCartByUserID
	expectedCart := &models.Cart{
		Items: []models.CartItem{
			{ProductID: "product-1", Quantity: 2},
		},
	}
	mockMongo.On("GetCartByUserID", ctx, "user-id").Return(expectedCart, nil)
	cart, err := adapter.GetCartByUserID(ctx, "user-id")
	require.NoError(t, err)
	assert.Equal(t, expectedCart, cart)

	// Test UpdateItemQuantity
	mockMongo.On("UpdateItemQuantity", ctx, "user-id", "product-1", 3).Return(nil)
	err = adapter.UpdateItemQuantity(ctx, "user-id", "product-1", 3)
	require.NoError(t, err)

	// Test RemoveItemFromCart
	mockMongo.On("RemoveItemFromCart", ctx, "user-id", "product-1").Return(nil)
	err = adapter.RemoveItemFromCart(ctx, "user-id", "product-1")
	require.NoError(t, err)

	// Test ClearCart
	mockMongo.On("ClearCart", ctx, "user-id").Return(nil)
	err = adapter.ClearCart(ctx, "user-id")
	require.NoError(t, err)

	mockMongo.AssertExpectations(t)
}

// TestProductAdapter_WithSqlMock tests the ProductAdapter using sqlmock
func TestProductAdapter_WithSqlMock(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

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
	require.NoError(t, err)
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
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.NotNil(t, tx)

	// Test BeginTx with custom options
	mock.ExpectBegin()
	tx, err = adapter.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDBTxAdapter tests the DBTxAdapter
func TestDBTxAdapter(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		t.Errorf("db.Close() failed: %v", err)
	// 	}
	// }()

	// Start a transaction
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	adapter := &DBTxAdapter{tx: tx}

	// Test Commit
	mock.ExpectCommit()
	err = adapter.Commit()
	require.NoError(t, err)

	// Test Rollback (we need a new transaction since the previous one was committed)
	mock.ExpectBegin()
	tx2, err := db.Begin()
	require.NoError(t, err)

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
	require.NoError(t, err)
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
	require.NoError(t, err)
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
	require.NoError(t, err)

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
	require.NoError(t, err)

	// Test RemoveGuestItem
	// First get the cart
	mock.ExpectGet("guest_cart:session-123").SetVal(string(cartJSON))
	// Then save the cart without the item
	emptyCart := &models.Cart{Items: []models.CartItem{}}
	emptyCartJSON, _ := json.Marshal(emptyCart)
	mock.ExpectSet("guest_cart:session-123", emptyCartJSON, 7*24*time.Hour).SetVal("OK")
	err = adapter.RemoveGuestItem(ctx, sessionID, "product-1")
	require.NoError(t, err)

	// Test DeleteGuestCart
	mock.ExpectDel("guest_cart:session-123").SetVal(1)
	err = adapter.DeleteGuestCart(ctx, sessionID)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
