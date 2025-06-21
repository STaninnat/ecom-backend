package paymenthandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/refund"
)

func (apicfg *HandlersPaymentConfig) HandlerRefundPayment(w http.ResponseWriter, r *http.Request, user database.User) {
	apicfg.SetupStripeAPI()

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"missing order id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	payment, err := apicfg.DB.GetPaymentByOrderID(ctx, orderID)
	if err != nil || payment.UserID != user.ID {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"payment not found",
			"Payment not found or unauthorized",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Payment not found")
		return
	}

	if payment.Status != "succeeded" {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"refund failed",
			"Payment cannot be refunded",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Payment cannot be refunded")
		return
	}

	if !payment.ProviderPaymentID.Valid {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing provider payment ID")
		return
	}

	refundParams := &stripe.RefundParams{
		PaymentIntent: stripe.String(payment.ProviderPaymentID.String),
	}
	_, err = refund.New(refundParams)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"stripe refund failed",
			"Stripe refund error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to process refund")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdatePaymentStatusByID(ctx, database.UpdatePaymentStatusByIDParams{
		ID:     payment.ID,
		Status: "cancelled",
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"update db failed",
			"Failed to update payment status",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update payment status")
		return
	}

	err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
		ID:        payment.OrderID,
		Status:    "cancelled",
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"update order failed",
			"Failed to update order status",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update order")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refund_payment",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "refund_payment", "Refund successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Refund processed"})
}
