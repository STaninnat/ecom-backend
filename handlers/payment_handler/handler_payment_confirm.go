package paymenthandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/refund"
)

func (apicfg *HandlersPaymentConfig) HandlerConfirmPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	apicfg.SetupStripeAPI()

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req ConfirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"invalid request",
			"Failed to decode body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	payment, err := apicfg.DB.GetPaymentByOrderID(ctx, req.OrderID)
	if err != nil || payment.UserID != user.ID {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"payment not found",
			"Payment not found or not owned by user",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Payment not found")
		return
	}

	intent, err := paymentintent.Get(payment.ProviderPaymentID.String, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"stripe error",
			"Failed to fetch payment intent",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to confirm payment")
		return
	}

	params := &stripe.RefundListParams{
		PaymentIntent: stripe.String(intent.ID),
	}

	refunds := refund.List(params)

	var newStatus string
	hasRefund := false

	for refunds.Next() {
		refund := refunds.Refund()
		if refund.Status == stripe.RefundStatusSucceeded {
			hasRefund = true
			break
		}
	}

	if hasRefund {
		newStatus = "refunded"
	} else {
		switch intent.Status {
		case stripe.PaymentIntentStatusSucceeded:
			newStatus = "succeeded"
		case stripe.PaymentIntentStatusCanceled:
			newStatus = "cancelled"
		case stripe.PaymentIntentStatusRequiresPaymentMethod, stripe.PaymentIntentStatusRequiresConfirmation:
			newStatus = "pending"
		default:
			newStatus = "failed"
		}
	}

	timeNow := time.Now().UTC()

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"start tx failed",
			"Transaction begin error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
		ID:        payment.ID,
		Status:    newStatus,
		UpdatedAt: timeNow,
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"update payment status failed",
			"Failed to update payment status",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update payment")
		return
	}

	statusMap := map[string]string{
		"succeeded": "paid",
		"refunded":  "refunded",
	}

	if updatedStatus, ok := statusMap[newStatus]; ok {
		err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
			ID:        payment.OrderID,
			Status:    updatedStatus,
			UpdatedAt: timeNow,
		})
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"confirm_payment",
				"update order failed",
				"Failed to update order status",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update order")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"confirm_payment",
			"commit failed",
			"Transaction commit error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction commit failed")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "confirm_payment", "Payment confirmation success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, ConfirmPaymentResponse{Status: newStatus})
}
