package cart_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/handlers/cart"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func setup() (*cart.HandlersCartConfig, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	cfg := &cart.HandlersCartConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				RedisClient: db,
			},
		},
	}

	return cfg, mock
}

func TestGetGuestCart_Success(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "session123"
	cartData := &models.Cart{Items: []models.CartItem{{ProductID: "p1", Quantity: 2}}}
	data, _ := json.Marshal(cartData)

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetVal(string(data))

	got, err := cfg.GetGuestCart(ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, cartData, got)
}

func TestGetGuestCart_NotFound(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "session123"

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).RedisNil()

	got, err := cfg.GetGuestCart(ctx, sessionID)
	assert.NoError(t, err)
	assert.Empty(t, got.Items)
}

func TestGetGuestCart_RedisError(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "session123"

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetErr(errors.New("redis error"))

	got, err := cfg.GetGuestCart(ctx, sessionID)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to get guest cart")
}

func TestSaveGuestCart_Success(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "session123"
	cartData := &models.Cart{Items: []models.CartItem{{ProductID: "p1", Quantity: 2}}}
	data, _ := json.Marshal(cartData)

	mock.ExpectSet(cart.GuestCartPrefix+sessionID, data, cart.TTL).SetVal("OK")

	err := cfg.SaveGuestCart(ctx, sessionID, cartData)
	assert.NoError(t, err)
}

func TestSaveGuestCart_FailRedisSet(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "session123"
	cartData := &models.Cart{Items: []models.CartItem{{ProductID: "p1", Quantity: 2}}}
	data, _ := json.Marshal(cartData)

	mock.ExpectSet(cart.GuestCartPrefix+sessionID, data, cart.TTL).SetErr(errors.New("redis set failed"))

	err := cfg.SaveGuestCart(ctx, sessionID, cartData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save guest cart to Redis")
}

func TestUpdateGuestItemQuantity_Success(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"
	productID := "p1"

	initial := &models.Cart{Items: []models.CartItem{{ProductID: productID, Quantity: 1}}}
	updated := &models.Cart{Items: []models.CartItem{{ProductID: productID, Quantity: 5}}}
	dataInitial, _ := json.Marshal(initial)
	dataUpdated, _ := json.Marshal(updated)

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetVal(string(dataInitial))
	mock.ExpectSet(cart.GuestCartPrefix+sessionID, []byte(dataUpdated), cart.TTL).SetVal("OK")

	err := cfg.UpdateGuestItemQuantity(ctx, sessionID, productID, 5)
	assert.NoError(t, err)
}

func TestUpdateGuestItemQuantity_GetCartError(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"
	productID := "p1"

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetErr(errors.New("redis down"))

	err := cfg.UpdateGuestItemQuantity(ctx, sessionID, productID, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis down")
}

func TestUpdateGuestItemQuantity_ItemNotFound(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"
	productID := "not-exist"

	initial := &models.Cart{Items: []models.CartItem{
		{ProductID: "p-other", Quantity: 1},
	}}
	dataInitial, _ := json.Marshal(initial)

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetVal(string(dataInitial))

	err := cfg.UpdateGuestItemQuantity(ctx, sessionID, productID, 3)
	assert.EqualError(t, err, "item not found")
}

func TestRemoveGuestItem_Success(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"

	initial := &models.Cart{Items: []models.CartItem{
		{ProductID: "p1", Quantity: 1},
		{ProductID: "p2", Quantity: 2},
	}}
	expected := &models.Cart{Items: []models.CartItem{
		{ProductID: "p2", Quantity: 2},
	}}

	dataInitial, _ := json.Marshal(initial)
	dataExpected, _ := json.Marshal(expected)

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetVal(string(dataInitial))
	mock.ExpectSet(cart.GuestCartPrefix+sessionID, []byte(dataExpected), cart.TTL).SetVal("OK")

	err := cfg.RemoveGuestItem(ctx, sessionID, "p1")
	assert.NoError(t, err)
}

func TestRemoveGuestItem_GetCartError(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetErr(errors.New("redis error"))

	err := cfg.RemoveGuestItem(ctx, sessionID, "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis error")
}

func TestRemoveGuestItem_SaveCartError(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"

	initial := &models.Cart{Items: []models.CartItem{
		{ProductID: "p1", Quantity: 1},
		{ProductID: "p2", Quantity: 2},
	}}
	expected := &models.Cart{Items: []models.CartItem{
		{ProductID: "p2", Quantity: 2},
	}}

	dataInitial, _ := json.Marshal(initial)
	dataExpected, _ := json.Marshal(expected)

	mock.ExpectGet(cart.GuestCartPrefix + sessionID).SetVal(string(dataInitial))
	mock.ExpectSet(cart.GuestCartPrefix+sessionID, []byte(dataExpected), cart.TTL).SetErr(errors.New("redis set failed"))

	err := cfg.RemoveGuestItem(ctx, sessionID, "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis set failed")
}

func TestDeleteGuestCart_Success(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"

	mock.ExpectDel(cart.GuestCartPrefix + sessionID).SetVal(1)

	err := cfg.DeleteGuestCart(ctx, sessionID)
	assert.NoError(t, err)
}

func TestDeleteGuestCart_Error(t *testing.T) {
	cfg, mock := setup()
	ctx := context.Background()
	sessionID := "s1"

	mock.ExpectDel(cart.GuestCartPrefix + sessionID).SetErr(errors.New("redis del failed"))

	err := cfg.DeleteGuestCart(ctx, sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis del failed")
}
