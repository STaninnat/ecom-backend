package orderhandlers

import (
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

type HandlersOrderConfig struct {
	*handlers.HandlersConfig
}

type CreateOrderRequest struct {
	Items             []OrderItemInput `json:"items"`
	PaymentMethod     string           `json:"payment_method"`
	ShippingAddress   string           `json:"shipping_address"`
	ContactPhone      string           `json:"contact_phone"`
	ExternalPaymentID string           `json:"external_payment_id,omitempty"`
	TrackingNumber    string           `json:"tracking_number,omitempty"`
}

type OrderItemInput struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
}

type UserOrderResponse struct {
	OrderID         string              `json:"order_id"`
	TotalAmount     string              `json:"total_amount"`
	Status          string              `json:"status"`
	PaymentMethod   string              `json:"payment_method,omitempty"`
	TrackingNumber  string              `json:"tracking_number,omitempty"`
	ShippingAddress string              `json:"shipping_address,omitempty"`
	ContactPhone    string              `json:"contact_phone,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	Items           []OrderItemResponse `json:"items"`
}

type OrderResponse struct {
	Message string `json:"message"`
	OrderID string `json:"order_id"`
}

type OrderDetailResponse struct {
	Order database.Order       `json:"order"`
	Items []database.OrderItem `json:"items"`
}
