package orderhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersOrderConfig) HandlerGetAllOrders(w http.ResponseWriter, r *http.Request, _ database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orders, err := apicfg.DB.ListAllOrders(ctx)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"list_all_orders",
			"db_error",
			"Failed to list orders",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Could not fetch orders")
		return
	}

	apicfg.LogHandlerSuccess(ctx, "list_all_orders", "Listed all orders", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, orders)
}

func (apicfg *HandlersOrderConfig) HandlerGetUserOrders(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orders, err := apicfg.DB.GetOrderByUserID(ctx, user.ID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_user_orders",
			"get order failed",
			"Failed to get orders",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	var response []UserOrderResponse
	for _, order := range orders {
		items, err := apicfg.DB.GetOrderItemsByOrderID(ctx, order.ID)
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"get_user_orders",
				"fetch items failed",
				"Failed to get order items",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch order items")
			return
		}

		var itemResponses []OrderItemResponse
		for _, item := range items {
			itemResponses = append(itemResponses, OrderItemResponse{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  int(item.Quantity),
				Price:     item.Price,
			})
		}

		response = append(response, UserOrderResponse{
			OrderID:         order.ID,
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			PaymentMethod:   order.PaymentMethod.String,
			TrackingNumber:  order.TrackingNumber.String,
			ShippingAddress: order.ShippingAddress.String,
			ContactPhone:    order.ContactPhone.String,
			CreatedAt:       order.CreatedAt,
			Items:           itemResponses,
		})
	}

	middlewares.RespondWithJSON(w, http.StatusOK, response)
}

func (apicfg *HandlersOrderConfig) HandlerGetOrderByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		apicfg.LogHandlerError(
			ctx,
			"get_order_by_id",
			"missing order id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order ID")
		return
	}

	order, err := apicfg.DB.GetOrderByID(ctx, orderID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_order_by_id",
			"order not found",
			"Order not found or DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	if order.UserID != user.ID && user.Role != "admin" {
		apicfg.LogHandlerError(
			ctx,
			"get_order_by_id",
			"unauthorized",
			"User is not authorized to view this order",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusForbidden, "Access Denied")
		return
	}

	items, err := apicfg.DB.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_order_by_id",
			"failed to fetch items",
			"Error getting order items",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch order items")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_order_by_id", "Fetched order details", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, OrderDetailResponse{
		Order: order,
		Items: items,
	})
}

func (apicfg *HandlersOrderConfig) HandlerGetOrderItemsByOrderID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		apicfg.LogHandlerError(
			ctx,
			"get_order_items",
			"missing order id",
			"Order ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing order_id")
		return
	}

	items, err := apicfg.DB.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_order_items",
			"fetch items failed",
			"Failed to fetch order items",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch order items")
		return
	}

	var response []OrderItemResponse
	for _, item := range items {
		response = append(response, OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
			Price:     item.Price,
		})
	}

	middlewares.RespondWithJSON(w, http.StatusOK, response)
}
