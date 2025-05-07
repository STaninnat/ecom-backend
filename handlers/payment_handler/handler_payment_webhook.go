package paymenthandlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func (apicfg *HandlersPaymentConfig) HandlerStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"payment_webhook",
			"read fialed",
			"Error reading req body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusServiceUnavailable, "Read error")
		return
	}

	endpointSecret := apicfg.StripeWebhookSecret
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"payment_webhook",
			"signature verification failed",
			"Error verificating signature",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Signature verification failed")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"payment_webhook",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	if event.Type == "payment_intent.succeeded" {
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			apicfg.LogHandlerError(
				ctx,
				"payment_webhook",
				"bad payment intent",
				"Bad payment intent",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusBadRequest, "Bad payment intent")
			return
		}
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "succeeded",
			ProviderPaymentID: utils.ToNullString(pi.ID),
		})
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"payment_webhook",
				"update payment failed",
				"Error updating payment",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update payment")
			return
		}
	}

	if event.Type == "payment_intent.refunded" {
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			apicfg.LogHandlerError(
				ctx,
				"payment_webhook",
				"bad payment intent",
				"Bad payment intent",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusBadRequest, "Bad payment intent")
			return
		}
		err = queries.UpdatePaymentStatusByProviderPaymentID(ctx, database.UpdatePaymentStatusByProviderPaymentIDParams{
			Status:            "refunded",
			ProviderPaymentID: utils.ToNullString(pi.ID),
		})
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"payment_webhook",
				"update payment failed",
				"Error updating payment",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update payment")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"payment_webhook",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{
		Message: "Updated payment successfully",
	})
}
