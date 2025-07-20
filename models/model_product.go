package models

import "time"

// Product represents an item available for purchase in the e-commerce system.
// It contains product details, pricing, inventory, and availability status.
type Product struct {
	ID          string    `json:"id"`          // Unique identifier for the product
	CategoryID  string    `json:"category_id"` // ID of the category this product belongs to
	Name        string    `json:"name"`        // Product name/title
	Description string    `json:"description"` // Detailed product description
	Price       float64   `json:"price"`       // Current selling price
	Stock       int32     `json:"stock"`       // Available quantity in inventory
	ImageURL    string    `json:"image_url"`   // URL to the product's main image
	IsActive    bool      `json:"is_active"`   // Whether the product is available for purchase
	CreatedAt   time.Time `json:"created_at"`  // When the product was added to the catalog
	UpdatedAt   time.Time `json:"updated_at"`  // When the product information was last updated
}
