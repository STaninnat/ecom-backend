package cart

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
)

type AddItemReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (apicfg *HandlersCartConfig) HandlerAddItemToUserCart(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req AddItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	product, err := apicfg.HandlersConfig.DB.GetProductByID(ctx, req.ProductID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"get product failed",
			"Error getting product info by id",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"convert format failed",
			"Error converting price format",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Invalid product price format")
		return
	}

	item := models.CartItem{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     price,
		Name:      product.Name,
	}

	if err := apicfg.CartMG.AddItemToCart(ctx, user.ID, item); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_to_cart",
			"failed to add item",
			"DB error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to add item")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "add_item_to_cart", "Added item to cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item added to cart",
	})
}

func (apicfg *HandlersCartConfig) HandlerAddItemToGuestCart(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"missing session ID",
			"Session ID not found in request",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req AddItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID and quantity are required")
		return
	}

	timeNow := time.Now().UTC()

	product, err := apicfg.HandlersConfig.DB.GetProductByID(ctx, req.ProductID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"get product failed",
			"Error to getting product",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"convert format failed",
			"Error converting price format",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Invalid product price format")
		return
	}

	cart, err := apicfg.GetGuestCart(ctx, sessionID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"get guest cart failed",
			"Error to getting guest cart",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve cart")
		return
	}
	if cart == nil {
		cart = &models.Cart{
			ID:        sessionID,
			UserID:    "",
			Items:     []models.CartItem{},
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}

	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == req.ProductID {
			cart.Items[i].Quantity += req.Quantity
			found = true
			break
		}
	}

	if !found {
		cart.Items = append(cart.Items, models.CartItem{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     price,
			Name:      product.Name,
		})
	}
	cart.UpdatedAt = timeNow

	if err := apicfg.SaveGuestCart(ctx, sessionID, cart); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"add_item_guest_cart",
			"save guest cart failed",
			"Error to saving guest cart",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to save cart")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "add_item_guest_cart", "Added item to guest cart", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Item added to guest cart",
	})
}
