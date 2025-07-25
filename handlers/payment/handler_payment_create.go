// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_payment_create.go: Payment intent creation handler with request validation and error handling.

// HandlerCreatePayment handles HTTP POST requests to create a new payment intent.
// @Summary      Create payment intent
// @Description  Creates a new payment intent for an order
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payment  body  object{}  true  "Payment intent payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/intent [post]
func (cfg *HandlersPaymentConfig) HandlerCreatePayment(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req CreatePaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"create_payment",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create payment using service
	result, err := cfg.GetPaymentService().CreatePayment(ctx, CreatePaymentParams{
		OrderID:  req.OrderID,
		UserID:   user.ID,
		Currency: req.Currency,
	})

	if err != nil {
		cfg.handlePaymentError(w, r, err, "create_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "create_payment", "Created payment successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, CreatePaymentIntentResponse{
		ClientSecret: result.ClientSecret,
	})
}
