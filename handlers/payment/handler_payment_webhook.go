// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"io"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
)

// handler_payment_webhook.go: Stripe webhook handler for payment event processing and validation.

// HandlerStripeWebhook handles HTTP POST requests from Stripe webhooks.
// Validates the webhook signature, processes the payload, and delegates handling to the payment service.
// On success, logs the event and responds with a confirmation; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the webhook payload
func (cfg *HandlersPaymentConfig) HandlerStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"payment_webhook",
			"read_failed",
			"Error reading request body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusServiceUnavailable, "Read error")
		return
	}

	// Validate content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		cfg.Logger.LogHandlerError(
			ctx,
			"payment_webhook",
			"invalid_content_type",
			"Expected application/json",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid content type")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"payment_webhook",
			"missing_signature",
			"Missing Stripe signature header",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing signature")
		return
	}

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Handle webhook using service
	err = paymentService.HandleWebhook(ctx, payload, signature, cfg.StripeWebhookSecret)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "payment_webhook", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "payment_webhook", "Webhook processed successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Updated payment successfully",
	})
}
