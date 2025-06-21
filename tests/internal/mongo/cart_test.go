package intmongo_test

import (
	"context"
	"errors"
	"testing"

	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// --- Mock Interfaces for Cart ---

type mockCartCollection struct {
	mock.Mock
}

func (m *mockCartCollection) FindOne(ctx context.Context, filter any) intmongo.CartSingleResultInterface {
	args := m.Called(ctx, filter)
	return args.Get(0).(intmongo.CartSingleResultInterface)
}

func (m *mockCartCollection) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

type mockCartSingleResult struct {
	mock.Mock
}

func (m *mockCartSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// --- Unit Tests ---

func TestGetCartByUserID_Found(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	mockRes := new(mockCartSingleResult)

	cart := &models.Cart{UserID: "user1", Items: []models.CartItem{{ProductID: "p1", Quantity: 2}}}

	mockCol.On("FindOne", ctx, mock.Anything).Return(mockRes)
	mockRes.On("Decode", mock.AnythingOfType("*models.Cart")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cart)
		*arg = *cart
	}).Return(nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	result, err := store.GetCartByUserID(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", result.UserID)
	assert.Len(t, result.Items, 1)
}

func TestGetCartByUserID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	mockRes := new(mockCartSingleResult)

	mockCol.On("FindOne", ctx, mock.Anything).Return(mockRes)
	mockRes.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)

	store := &intmongo.CartMongo{Collection: mockCol}
	result, err := store.GetCartByUserID(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", result.UserID)
	assert.Empty(t, result.Items)
}

func TestAddItemToCart_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	item := models.CartItem{ProductID: "p1", Quantity: 2}

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.AddItemToCart(ctx, "user1", item)

	assert.NoError(t, err)
}

func TestAddItemToCart_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	item := models.CartItem{ProductID: "p1", Quantity: 2}

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("update failed"))

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.AddItemToCart(ctx, "user1", item)

	assert.Error(t, err)
}

func TestRemoveItemFromCart_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.RemoveItemFromCart(ctx, "user1", "p1")

	assert.NoError(t, err)
}

func TestRemoveItemFromCart_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("update error"))

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.RemoveItemFromCart(ctx, "user1", "p1")

	assert.Error(t, err)
}

func TestClearCart_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.ClearCart(ctx, "user1")

	assert.NoError(t, err)
}

func TestClearCart_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("clear error"))

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.ClearCart(ctx, "user1")

	assert.Error(t, err)
}

func TestUpdateItemQuantity_Update(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.UpdateItemQuantity(ctx, "user1", "p1", 3)

	assert.NoError(t, err)
}

func TestUpdateItemQuantity_Remove(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.UpdateItemQuantity(ctx, "user1", "p1", 0)

	assert.NoError(t, err)
}

func TestUpdateItemQuantity_NotMatched(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{MatchedCount: 0}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	err := store.UpdateItemQuantity(ctx, "user1", "p1", 5)

	assert.EqualError(t, err, "item not found in cart")
}

func TestUpsertCart_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	cart := models.Cart{UserID: "user1", Items: []models.CartItem{{ProductID: "p1", Quantity: 1}}}
	err := store.UpsertCart(ctx, "user1", cart)

	assert.NoError(t, err)
}

func TestUpsertCart_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("upsert error"))

	store := &intmongo.CartMongo{Collection: mockCol}
	cart := models.Cart{UserID: "user1", Items: []models.CartItem{}}
	err := store.UpsertCart(ctx, "user1", cart)

	assert.Error(t, err)
}

func TestMergeGuestCartToUser_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	mockRes := new(mockCartSingleResult)

	existingCart := &models.Cart{
		UserID: "user1",
		Items:  []models.CartItem{{ProductID: "p1", Quantity: 2}},
	}

	mockCol.On("FindOne", ctx, mock.Anything).Return(mockRes)
	mockRes.On("Decode", mock.AnythingOfType("*models.Cart")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cart)
		*arg = *existingCart
	}).Return(nil)

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{}, nil)

	store := &intmongo.CartMongo{Collection: mockCol}
	items := []models.CartItem{{ProductID: "p1", Quantity: 1}, {ProductID: "p2", Quantity: 1}}
	err := store.MergeGuestCartToUser(ctx, "user1", items)

	assert.NoError(t, err)
}

func TestMergeGuestCartToUser_Error_UpsertCart(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	mockRes := new(mockCartSingleResult)

	// Return valid cart
	existingCart := &models.Cart{
		UserID: "user1",
		Items:  []models.CartItem{{ProductID: "p1", Quantity: 1}},
	}

	mockCol.On("FindOne", ctx, mock.Anything).Return(mockRes)
	mockRes.On("Decode", mock.AnythingOfType("*models.Cart")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cart)
		*arg = *existingCart
	}).Return(nil)

	// Simulate UpsertCart error
	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("upsert failed"))

	store := &intmongo.CartMongo{Collection: mockCol}
	items := []models.CartItem{{ProductID: "p2", Quantity: 2}}

	err := store.MergeGuestCartToUser(ctx, "user1", items)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upsert failed")
}

func TestMergeGuestCartToUser_Error_GetCart(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCartCollection)
	mockRes := new(mockCartSingleResult)

	mockCol.On("FindOne", ctx, mock.Anything).Return(mockRes)
	mockRes.On("Decode", mock.Anything).Return(errors.New("db error"))

	store := &intmongo.CartMongo{Collection: mockCol}
	items := []models.CartItem{{ProductID: "p1", Quantity: 1}}

	err := store.MergeGuestCartToUser(ctx, "user1", items)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")

	// ยืนยันว่า UpdateOne ไม่ถูกเรียก
	mockCol.AssertNotCalled(t, "UpdateOne", mock.Anything, mock.Anything, mock.Anything)
}
