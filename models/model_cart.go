// Package models defines data structures and database models for the ecom-backend project.
package models

import "time"

// model_cart.go: Defines Cart and CartItem models for shopping cart functionality.

// CartItem represents a single item in a user's shopping cart.
// It contains product information and quantity for checkout purposes.
type CartItem struct {
	ProductID string  `bson:"product_id" json:"product_id"` // ID of the product in the cart
	Quantity  int     `bson:"quantity" json:"quantity"`     // Number of units of this product
	Price     float64 `bson:"price" json:"price"`           // Current price per unit
	Name      string  `bson:"name" json:"name"`             // Product name for display
}

// Cart represents a user's shopping cart containing multiple items.
// It tracks the user's selected products before checkout.
type Cart struct {
	ID        string     `bson:"_id,omitempty" json:"id"`      // Unique identifier for the cart
	UserID    string     `bson:"user_id" json:"user_id"`       // ID of the user who owns this cart
	Items     []CartItem `bson:"items" json:"items"`           // List of items in the cart
	CreatedAt time.Time  `bson:"created_at" json:"created_at"` // When the cart was created
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"` // When the cart was last modified
}
