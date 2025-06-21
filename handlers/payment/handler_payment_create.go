package paymenthandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
)

func (apicfg *HandlersPaymentConfig) HandlerCreatePayment(w http.ResponseWriter, r *http.Request, user database.User) {
	apicfg.SetupStripeAPI()

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req CreatePaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	order, err := apicfg.DB.GetOrderByID(ctx, req.OrderID)
	if err != nil || order.UserID != user.ID {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"order not found",
			"Order not found or DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Order not found")
		return
	}
	if order.Status != "pending" {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"order status invalid",
			"Error order status isn't equal pending",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Order already paid or invalid")
		return
	}

	totalAmount, err := strconv.ParseFloat(order.TotalAmount, 64)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"invalid total amount",
			"Failed to parse total amount as float",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid total amount")
		return
	}
	amountInSatang := int64(totalAmount * 100)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInSatang),
		Currency: stripe.String(req.Currency),
		Metadata: map[string]string{
			"order_id": order.ID,
			"user_id":  user.ID,
		},
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"create payment intent failed",
			"Error creating payment intent",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create payment intent")
		return
	}

	paymentID := uuid.New().String()
	timeNow := time.Now().UTC()

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	_, err = queries.CreatePayment(ctx, database.CreatePaymentParams{
		ID:                paymentID,
		OrderID:           order.ID,
		UserID:            user.ID,
		Amount:            order.TotalAmount,
		Currency:          req.Currency,
		Status:            "created",
		Provider:          "stripe",
		ProviderPaymentID: utils.ToNullString(intent.ID),
		CreatedAt:         timeNow,
		UpdatedAt:         timeNow,
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"record payment failed",
			"Error recording payment",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to record payment")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_payment",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "create_payment", "Created payment successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, CreatePaymentIntentResponse{
		ClientSecret: intent.ClientSecret,
	})
}
