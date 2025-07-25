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

// handler_payment_confirm.go: HTTP handler for processing payment confirmation and responding with status.

// HandlerConfirmPayment handles HTTP POST requests to confirm a payment.
// @Summary      Confirm payment
// @Description  Confirms a payment for an order
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payment  body  object{}  true  "Payment confirmation payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/confirm [post]
func (cfg *HandlersPaymentConfig) HandlerConfirmPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req ConfirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"confirm_payment",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Confirm payment using service
	result, err := cfg.GetPaymentService().ConfirmPayment(ctx, ConfirmPaymentParams{
		OrderID: req.OrderID,
		UserID:  user.ID,
	})

	if err != nil {
		cfg.handlePaymentError(w, r, err, "confirm_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "confirm_payment", "Payment confirmation success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, ConfirmPaymentResponse{Status: result.Status})
}
