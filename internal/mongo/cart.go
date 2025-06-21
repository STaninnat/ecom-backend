package intmongo

import (
	"context"
	"fmt"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// --- CartMongo ---
type CartMongo struct {
	Collection CartCollectionInterface
}

func NewCartMongo(db *mongo.Database) *CartMongo {
	return &CartMongo{
		Collection: &MongoCartCollectionAdapter{
			Inner: db.Collection("carts"),
		},
	}
}

func (c *CartMongo) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	timeNow := time.Now().UTC()

	filter := bson.M{"user_id": userID}

	var cart models.Cart
	err := c.Collection.FindOne(ctx, filter).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &models.Cart{
				UserID:    userID,
				Items:     []models.CartItem{},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			}, nil
		}
		return nil, err
	}

	return &cart, nil
}

func (c *CartMongo) AddItemToCart(ctx context.Context, userID string, item models.CartItem) error {
	timeNow := time.Now().UTC()

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"created_at": timeNow,
		},
		"$set": bson.M{
			"updated_at": timeNow,
		},
		"$push": bson.M{
			"items": item,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := c.Collection.UpdateOne(ctx, filter, update, opts)

	return err
}

func (c *CartMongo) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{"product_id": productID},
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := c.Collection.UpdateOne(ctx, filter, update)

	return err
}

func (c *CartMongo) ClearCart(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"items":      []models.CartItem{},
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := c.Collection.UpdateOne(ctx, filter, update)

	return err
}

func (c *CartMongo) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) error {
	filter := bson.M{
		"user_id":          userID,
		"items.product_id": productID,
	}

	var update bson.M
	if quantity <= 0 {
		update = bson.M{
			"$pull": bson.M{
				"items": bson.M{"product_id": productID},
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				"items.$.quantity": quantity,
				"updated_at":       time.Now().UTC(),
			},
		}
	}

	result, err := c.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("item not found in cart")
	}

	return nil
}

func (c *CartMongo) UpsertCart(ctx context.Context, userID string, cart models.Cart) error {
	timeNow := time.Now().UTC()

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"items":      cart.Items,
			"updated_at": timeNow,
		},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"created_at": timeNow,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := c.Collection.UpdateOne(ctx, filter, update, opts)

	return err
}

func (c *CartMongo) MergeGuestCartToUser(ctx context.Context, userID string, items []models.CartItem) error {
	cart, err := c.GetCartByUserID(ctx, userID)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	// map for deduplication and merge
	itemMap := map[string]models.CartItem{}
	for _, item := range items {
		itemMap[item.ProductID] = item
	}

	for _, existing := range cart.Items {
		if guestItem, ok := itemMap[existing.ProductID]; ok {
			existing.Quantity += guestItem.Quantity
			itemMap[existing.ProductID] = existing
		} else {
			itemMap[existing.ProductID] = existing
		}
	}

	mergedItems := []models.CartItem{}
	for _, v := range itemMap {
		mergedItems = append(mergedItems, v)
	}

	cart.Items = mergedItems
	cart.UpdatedAt = time.Now().UTC()

	return c.UpsertCart(ctx, userID, *cart)
}
