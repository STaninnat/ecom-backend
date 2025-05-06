package producthandlers

import (
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/utils"
)

type HandlersProductConfig struct {
	*handlers.HandlersConfig
}

type ProductRequest struct {
	ID          string  `json:"id,omitempty"`
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type FilterProductsRequest struct {
	CategoryID utils.NullString  `json:"category_id,omitempty"`
	IsActive   utils.NullBool    `json:"is_active,omitempty"`
	MinPrice   utils.NullFloat64 `json:"min_price,omitempty"`
	MaxPrice   utils.NullFloat64 `json:"max_price,omitempty"`
}

type CategoryWithIDRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type productResponse struct {
	Message   string `json:"message"`
	ProductID string `json:"product_id"`
}
