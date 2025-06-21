package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"github.com/redis/go-redis/v9"
)

var TTL = 7 * 24 * time.Hour

const GuestCartPrefix = "guest_cart:"

func (apicfg *HandlersCartConfig) GetGuestCart(ctx context.Context, sessionTD string) (*models.Cart, error) {
	key := GuestCartPrefix + sessionTD

	val, err := apicfg.HandlersConfig.RedisClient.Get(ctx, key).Result()
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

func (apicfg *HandlersCartConfig) SaveGuestCart(ctx context.Context, sessionID string, cart *models.Cart) error {
	key := GuestCartPrefix + sessionID

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	err = apicfg.HandlersConfig.RedisClient.Set(ctx, key, data, TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to save guest cart to Redis: %w", err)
	}

	return nil
}

func (apicfg *HandlersCartConfig) UpdateGuestItemQuantity(ctx context.Context, sessionID, productID string, quantity int) error {
	cart, err := apicfg.GetGuestCart(ctx, sessionID)
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

	return apicfg.SaveGuestCart(ctx, sessionID, cart)
}

func (apicfg *HandlersCartConfig) RemoveGuestItem(ctx context.Context, sessionID, productID string) error {
	cart, err := apicfg.GetGuestCart(ctx, sessionID)
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

	return apicfg.SaveGuestCart(ctx, sessionID, cart)
}

func (apicfg *HandlersCartConfig) DeleteGuestCart(ctx context.Context, sessionID string) error {
	key := GuestCartPrefix + sessionID

	return apicfg.HandlersConfig.RedisClient.Del(ctx, key).Err()
}
