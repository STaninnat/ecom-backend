// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_payment_refund.go: Refund payment request handler delegating to payment service.

// HandlerRefundPayment handles HTTP POST requests to process a payment refund.
// @Summary      Refund payment
// @Description  Processes a payment refund for a specific order
// @Tags         payments
// @Produce      json
// @Param        order_id  path  string  true  "Order ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/{order_id}/refund [post]
func (cfg *HandlersPaymentConfig) HandlerRefundPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"refund_payment",
			"missing_order_id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	// Process refund using service
	err := cfg.GetPaymentService().RefundPayment(ctx, RefundPaymentParams{
		OrderID: orderID,
		UserID:  user.ID,
	})

	if err != nil {
		cfg.handlePaymentError(w, r, err, "refund_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "refund_payment", "Refund successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Refund processed"})
}
