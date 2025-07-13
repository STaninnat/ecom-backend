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

// HandlerGetPayment handles HTTP GET requests to retrieve payment information for a specific order.
// It extracts the order ID from the URL parameters, validates it, and delegates retrieval to the payment service.
// On success, it logs the event and responds with the payment details; on error, it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
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

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Get payment using service
	result, err := paymentService.GetPayment(ctx, orderID, user.ID)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "get_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_payment", "Get Payment success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, *result)
}

// HandlerGetPaymentHistory handles HTTP GET requests to retrieve payment history for the authenticated user.
// It delegates retrieval to the payment service and returns the user's payment history.
// On success, it logs the event and responds with the payment history; on error, it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersPaymentConfig) HandlerGetPaymentHistory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Get payment history using service
	payments, err := paymentService.GetPaymentHistory(ctx, user.ID)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "get_history_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_history_payment", "Get Payment history success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, payments)
}

// HandlerAdminGetPayments handles HTTP GET requests to retrieve all payments for admin users.
// It extracts optional status filter from URL parameters and delegates retrieval to the payment service.
// On success, it logs the event and responds with the payment list; on error, it logs and returns the appropriate error response.
//
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated admin user
func (cfg *HandlersPaymentConfig) HandlerAdminGetPayments(w http.ResponseWriter, r *http.Request, _ database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	status := chi.URLParam(r, "status")

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Get all payments using service
	payments, err := paymentService.GetAllPayments(ctx, status)
	if err != nil {
		cfg.handlePaymentError(w, r, err, "admin_get_payments", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "admin_get_payments", "Get all payments success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, payments)
}
