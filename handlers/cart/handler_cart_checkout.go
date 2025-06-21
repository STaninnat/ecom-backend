package cart

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
)

func (apicfg *HandlersCartConfig) HandlerCheckoutCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()
	userID := user.ID

	cart, err := apicfg.CartMG.GetCartByUserID(ctx, userID)
	if err != nil || len(cart.Items) == 0 {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"checkout_cart",
			"empty or fetch error",
			"Error retrieving cart",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Cart is empty or cannot be retrieved")
		return
	}

	tx, err := apicfg.HandlersConfig.DBConn.BeginTx(ctx, nil)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"checkout_cart",
			"start tx failed",
			"Error starting transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	queries := apicfg.HandlersConfig.DB.WithTx(tx)
	totalAmount := 0.0
	timeNow := time.Now().UTC()

	for _, item := range cart.Items {
		qty32, err := safeIntToInt32(item.Quantity)
		if err != nil {
			middlewares.RespondWithError(w, http.StatusBadRequest, "Quantity too large")
			return
		}

		product, err := queries.GetProductByID(ctx, item.ProductID)
		if err != nil {
			apicfg.HandlersConfig.LogHandlerError(
				ctx,
				"checkout_cart",
				"query failed",
				"Error to fetch products",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusNotFound, "Product not found: "+item.ProductID)
			return
		}

		if product.Stock < qty32 {
			apicfg.HandlersConfig.LogHandlerError(
				ctx,
				"checkout_cart",
				"insufficient stock",
				"Insufficient stock for product",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusBadRequest, "Insufficient stock for product: "+item.ProductID)
			return
		}

		totalAmount += item.Price * float64(item.Quantity)
	}

	orderID := uuid.New().String()
	_, err = queries.CreateOrder(ctx, database.CreateOrderParams{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: fmt.Sprintf("%.2f", totalAmount),
		Status:      "pending",
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"checkout_cart",
			"create order failed",
			"Error to creating order",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	for _, item := range cart.Items {
		qty32, err := safeIntToInt32(item.Quantity)
		if err != nil {
			middlewares.RespondWithError(w, http.StatusBadRequest, "Quantity too large")
			return
		}

		negStock, err := safeIntToInt32(-item.Quantity)
		if err != nil {
			middlewares.RespondWithError(w, http.StatusBadRequest, "Quantity too large")
			return
		}

		err = queries.UpdateProductStock(ctx, database.UpdateProductStockParams{
			ID:    item.ProductID,
			Stock: negStock,
		})
		if err != nil {
			apicfg.HandlersConfig.LogHandlerError(
				ctx,
				"checkout_cart",
				"update stock failed",
				"Error to update stock",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update stock")
			return
		}

		err = queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  qty32,
			Price:     fmt.Sprintf("%.2f", item.Price),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
		if err != nil {
			apicfg.HandlersConfig.LogHandlerError(
				ctx,
				"checkout_cart",
				"create order item failed",
				"Error to creating order item",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create order item")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"checkout_cart",
			"commit tx failed",
			"Error committing transaction",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	if err := apicfg.CartMG.ClearCart(ctx, userID); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"checkout_cart",
			"clear cart failed",
			"Error to clear cart",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Order placed, but failed to clear cart")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "checkout_cart", "Order created successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, CartResponse{
		Message: "Order placed successfully",
		OrderID: orderID,
	})
}

func safeIntToInt32(i int) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, fmt.Errorf("value %d overflows int32", i)
	}

	return int32(i), nil
}
