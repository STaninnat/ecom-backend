package orderhandlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersOrderConfig) HandlerUpdateOrderStatus(w http.ResponseWriter, r *http.Request, user database.User) {
	type UpdateOrderStatusRequest struct {
		Status string `json:"status"`
	}

	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}

	if !validStatuses[req.Status] {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"invalid status",
			"Received invalid order status",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid order status")
		return
	}

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"missing order id",
			"Order ID must be provided",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	err = queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{
		ID:        orderID,
		Status:    req.Status,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"update failed",
			"Failed to update order status",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	err = tx.Commit()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"update_order_status",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	apicfg.LogHandlerSuccess(ctx, "update_order_status", "Order status updated successfully", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Order status updated successfully",
	})
}
