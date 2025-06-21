package paymenthandlers

import (
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/stripe/stripe-go/v82"
)

type HandlersPaymentConfig struct {
	*handlers.HandlersConfig
}

func (apicfg *HandlersPaymentConfig) SetupStripeAPI() {
	stripe.Key = apicfg.StripeSecretKey
}

type CreatePaymentIntentRequest struct {
	OrderID  string `json:"order_id"`
	Currency string `json:"currency"`
}

type CreatePaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
}

type ConfirmPaymentRequest struct {
	OrderID string `json:"order_id"`
}

type ConfirmPaymentResponse struct {
	Status string `json:"status"`
}

type GetPaymentResponse struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	UserID            string    `json:"user_id"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type PaymentHistoryItem struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	Amount            string    `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
