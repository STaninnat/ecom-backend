package paymenthandlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersPaymentConfig) HandlerGetPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	apicfg.SetupStripeAPI()

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		apicfg.LogHandlerError(
			ctx,
			"get_payment",
			"missing order id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	payment, err := apicfg.DB.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_payment",
			"payment not found",
			"Payment not found for the order",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Payment not found")
		return
	}

	if payment.UserID != user.ID {
		apicfg.LogHandlerError(
			ctx,
			"get_payment",
			"unauthorized",
			"User does not own this payment",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusForbidden, "Unauthorized")
		return
	}

	amount, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_payment",
			"invalid_amount",
			"Failed to parse payment amount",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Invalid payment amount")
		return
	}

	getPaymentResp := GetPaymentResponse{
		ID:                payment.ID,
		OrderID:           payment.OrderID,
		UserID:            payment.UserID,
		Amount:            amount,
		Currency:          payment.Currency,
		Status:            payment.Status,
		Provider:          payment.Provider,
		ProviderPaymentID: payment.ProviderPaymentID.String,
		CreatedAt:         payment.CreatedAt,
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_payment", "Get Payment success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, getPaymentResp)
}

func (apicfg *HandlersPaymentConfig) HandlerGetPaymentHistory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	payments, err := apicfg.DB.GetPaymentsByUserID(ctx, user.ID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_history_payment",
			"payment not found",
			"Payment not found for the user",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch payment history")
		return
	}

	resp := make([]PaymentHistoryItem, 0, len(payments))
	for _, p := range payments {
		resp = append(resp, PaymentHistoryItem{
			ID:                p.ID,
			OrderID:           p.OrderID,
			Amount:            p.Amount,
			Currency:          p.Currency,
			Status:            p.Status,
			Provider:          p.Provider,
			ProviderPaymentID: p.ProviderPaymentID.String,
			CreatedAt:         p.CreatedAt,
		})
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_history_payment", "Get Payment success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, resp)
}

func (apicfg *HandlersPaymentConfig) HandlerAdminGetPayments(w http.ResponseWriter, r *http.Request, _ database.User) {
	ctx := r.Context()
	ip, userAgent := handlers.GetRequestMetadata(r)

	status := chi.URLParam(r, "status")

	var payments []database.Payment
	var err error

	if status == "all" {
		payments, err = apicfg.DB.GetAllPayments(ctx)
	} else {
		payments, err = apicfg.DB.GetPaymentsByStatus(ctx, status)
	}

	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"admin_get_payments",
			"db query failed",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}

	resp := make([]PaymentHistoryItem, 0, len(payments))
	for _, p := range payments {
		resp = append(resp, PaymentHistoryItem{
			ID:                p.ID,
			OrderID:           p.OrderID,
			Amount:            p.Amount,
			Currency:          p.Currency,
			Status:            p.Status,
			Provider:          p.Provider,
			ProviderPaymentID: p.ProviderPaymentID.String,
			CreatedAt:         p.CreatedAt,
		})
	}

	apicfg.LogHandlerSuccess(ctx, "get_history_payment", "Get Payment success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, resp)
}
