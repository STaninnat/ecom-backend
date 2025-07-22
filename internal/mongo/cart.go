// Package mongo provides MongoDB repositories and helpers for the ecom-backend project.
package intmongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// cart.go: MongoDB repository and operations for shopping cart management.

// CartMongo handles cart operations in MongoDB.
type CartMongo struct {
	Collection CollectionInterface
}

// NewCartMongo creates a new CartMongo instance for the given MongoDB database.
func NewCartMongo(db *mongo.Database) *CartMongo {
	return &CartMongo{
		Collection: &MongoCollectionAdapter{
			Inner: db.Collection("carts"),
		},
	}
}

// GetCartByUserID retrieves a cart by user ID, creating an empty cart if not found.
func (c *CartMongo) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	timeNow := time.Now().UTC()
	filter := bson.M{"user_id": userID}

	var cart models.Cart
	result := c.Collection.FindOne(ctx, filter)
	err := result.Decode(&cart)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &models.Cart{
				ID:        generateCartID(userID), // Ensure ID is a string
				UserID:    userID,
				Items:     []models.CartItem{},
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			}, nil
		}
		return nil, fmt.Errorf("failed to decode cart: %w", err)
	}

	return &cart, nil
}

// generateCartID generates a unique string ID for a cart (simple version using userID and timestamp)
func generateCartID(userID string) string {
	return fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
}

// GetCartsByUserIDs retrieves multiple carts by user IDs.
func (c *CartMongo) GetCartsByUserIDs(ctx context.Context, userIDs []string) ([]*models.Cart, error) {
	if len(userIDs) == 0 {
		return []*models.Cart{}, nil
	}

	filter := bson.M{"user_id": bson.M{"$in": userIDs}}

	cursor, err := c.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find carts: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error or handle it as appropriate
			fmt.Printf("failed to close cursor: %v\n", err)
		}
	}()

	var carts []*models.Cart
	if err := cursor.All(ctx, &carts); err != nil {
		return nil, fmt.Errorf("failed to decode carts: %w", err)
	}

	return carts, nil
}

// AddItemToCart adds an item to a user's cart, creating the cart if it doesn't exist.
func (c *CartMongo) AddItemToCart(ctx context.Context, userID string, item models.CartItem) error {
	timeNow := time.Now().UTC()
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":        generateCartID(userID), // Ensure _id is a string
			"created_at": timeNow,
			"user_id":    userID, // Also ensure user_id is set on insert
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
	if err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	return nil
}

// AddItemsToCart adds multiple items to a user's cart.
func (c *CartMongo) AddItemsToCart(ctx context.Context, userID string, items []models.CartItem) error {
	if len(items) == 0 {
		return fmt.Errorf("items slice cannot be empty")
	}

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
			"items": bson.M{"$each": items},
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := c.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to add items to cart: %w", err)
	}

	return nil
}

// RemoveItemFromCart removes an item from a user's cart.
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
	if err != nil {
		return fmt.Errorf("failed to remove item from cart: %w", err)
	}

	return nil
}

// RemoveItemsFromCart removes multiple items from a user's cart.
func (c *CartMongo) RemoveItemsFromCart(ctx context.Context, userID string, productIDs []string) error {
	if len(productIDs) == 0 {
		return fmt.Errorf("product IDs slice cannot be empty")
	}

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{"product_id": bson.M{"$in": productIDs}},
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := c.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove items from cart: %w", err)
	}

	return nil
}

// ClearCart removes all items from a user's cart.
func (c *CartMongo) ClearCart(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"items":      []models.CartItem{},
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := c.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// ClearCarts removes all items from multiple users' carts.
func (c *CartMongo) ClearCarts(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("user IDs slice cannot be empty")
	}

	filter := bson.M{"user_id": bson.M{"$in": userIDs}}
	update := bson.M{
		"$set": bson.M{
			"items":      []models.CartItem{},
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := c.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear carts: %w", err)
	}

	return nil
}

// UpdateItemQuantity updates the quantity of an item in a user's cart.
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
		return fmt.Errorf("failed to update item quantity: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("item not found in cart")
	}

	return nil
}

// UpdateItemQuantities updates quantities of multiple items in a user's cart.
func (c *CartMongo) UpdateItemQuantities(ctx context.Context, userID string, updates map[string]int) error {
	if len(updates) == 0 {
		return fmt.Errorf("updates map cannot be empty")
	}

	// Process each update individually for simplicity and reliability
	for productID, quantity := range updates {
		err := c.UpdateItemQuantity(ctx, userID, productID, quantity)
		if err != nil {
			return fmt.Errorf("failed to update item %s: %w", productID, err)
		}
	}

	return nil
}

// UpsertCart creates or updates a user's cart.
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
	if err != nil {
		return fmt.Errorf("failed to upsert cart: %w", err)
	}

	return nil
}

// GetCartStats gets statistics about carts (total carts, total items, average items per cart).
func (c *CartMongo) GetCartStats(ctx context.Context) (map[string]any, error) {
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":             nil,
			"totalCarts":      bson.M{"$sum": 1},
			"totalItems":      bson.M{"$sum": bson.M{"$size": "$items"}},
			"avgItemsPerCart": bson.M{"$avg": bson.M{"$size": "$items"}},
		}},
		{"$project": bson.M{
			"_id":             0,
			"totalCarts":      1,
			"totalItems":      1,
			"avgItemsPerCart": bson.M{"$round": []any{"$avgItemsPerCart", 2}},
		}},
	}

	cursor, err := c.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate cart stats: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error or handle it as appropriate
			fmt.Printf("failed to close cursor: %v\n", err)
		}
	}()

	var results []map[string]any
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	if len(results) == 0 {
		return map[string]any{
			"totalCarts":      0,
			"totalItems":      0,
			"avgItemsPerCart": 0.0,
		}, nil
	}

	return results[0], nil
}

// DeleteCart deletes a user's cart completely.
func (c *CartMongo) DeleteCart(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID}

	result, err := c.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("cart not found")
	}

	return nil
}
