package orderhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
)

func (apicfg *HandlersOrderConfig) HandlerCreateOrder(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_order",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	orderID := uuid.New().String()
	timeNow := time.Now().UTC()

	var totalAmount float64
	for _, item := range params.Items {
		totalAmount += float64(item.Quantity) * item.Price
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_order",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	_, err = queries.CreateOrder(ctx, database.CreateOrderParams{
		ID:                orderID,
		UserID:            user.ID,
		TotalAmount:       fmt.Sprintf("%.2f", totalAmount),
		Status:            "pending",
		PaymentMethod:     utils.ToNullString(params.PaymentMethod),
		ExternalPaymentID: utils.ToNullString(params.ExternalPaymentID),
		TrackingNumber:    utils.ToNullString(params.TrackingNumber),
		ShippingAddress:   utils.ToNullString(params.ShippingAddress),
		ContactPhone:      utils.ToNullString(params.ContactPhone),
		CreatedAt:         timeNow,
		UpdatedAt:         timeNow,
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_order",
			"create order failed",
			"Error creating order",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	for _, item := range params.Items {
		err := queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  int32(item.Quantity),
			Price:     fmt.Sprintf("%.2f", item.Price),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"create_order",
				"create order item failed",
				"Error creating order item",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create order item")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_order",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "create_order", "Created order successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, OrderResponse{
		Message: "Created order successful",
		OrderID: orderID,
	})
}
