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

// handler_payment_get.go: Payment handlers for fetching payment info, history, and admin listing.

// HandlerGetPayment handles HTTP GET requests to retrieve payment information for a specific order.
// @Summary      Get payment by order ID
// @Description  Retrieves payment information for a specific order
// @Tags         payments
// @Produce      json
// @Param        order_id  path  string  true  "Order ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/{order_id} [get]
func (cfg *HandlersPaymentConfig) HandlerGetPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"get_payment",
			"missing_order_id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	// Get payment using service
	result, err := cfg.GetPaymentService().GetPayment(ctx, orderID, user.ID)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "get_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_payment", "Get Payment success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, *result)
}

// HandlerGetPaymentHistory handles HTTP GET requests to retrieve payment history for the authenticated user.
// @Summary      Get payment history
// @Description  Retrieves payment history for the authenticated user
// @Tags         payments
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/history [get]
func (cfg *HandlersPaymentConfig) HandlerGetPaymentHistory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get payment history using service
	payments, err := cfg.GetPaymentService().GetPaymentHistory(ctx, user.ID)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "get_history_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_history_payment", "Get Payment history success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, payments)
}

// HandlerAdminGetPayments handles HTTP GET requests to retrieve all payments for admin users.
// @Summary      Admin get payments by status
// @Description  Retrieves all payments filtered by status (admin only)
// @Tags         payments
// @Produce      json
// @Param        status  path  string  true  "Payment status"
// @Success      200  {array}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /v1/payments/admin/{status} [get]
func (cfg *HandlersPaymentConfig) HandlerAdminGetPayments(w http.ResponseWriter, r *http.Request, _ database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	status := chi.URLParam(r, "status")

	// Get all payments using service
	payments, err := cfg.GetPaymentService().GetAllPayments(ctx, status)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "admin_get_payments", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "admin_get_payments", "Get all payments success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, payments)
}
