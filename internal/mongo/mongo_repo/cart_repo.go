package mongorepo

import (
	"context"

	"github.com/STaninnat/ecom-backend/models"
)

type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	AddItemToCart(ctx context.Context, userID string, item models.CartItem) error
	RemoveItemFromCart(ctx context.Context, userID string, productID string) error
	ClearCart(ctx context.Context, userID string) error
}
