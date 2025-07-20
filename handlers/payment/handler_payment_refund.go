package paymenthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

// HandlerRefundPayment handles HTTP POST requests to process a payment refund.
// Extracts the order ID from URL parameters, validates it, and delegates refund processing to the payment service.
// On success, logs the event and responds with a confirmation message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
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

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Process refund using service
	err := paymentService.RefundPayment(ctx, RefundPaymentParams{
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
