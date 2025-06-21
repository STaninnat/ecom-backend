package models

import "time"

type CartItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Price     float64 `bson:"price" json:"price"`
	Name      string  `bson:"name" json:"name"`
}

type Cart struct {
	ID        string     `bson:"_id,omitempty" json:"id"`
	UserID    string     `bson:"user_id" json:"user_id"`
	Items     []CartItem `bson:"items" json:"items"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
}
