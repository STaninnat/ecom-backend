package producthandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersProductConfig) HandlerGetAllProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var (
		products []database.Product
		err      error
	)

	if user.Role == "admin" {
		products, err = apicfg.DB.GetAllProducts(ctx)
	} else {
		products, err = apicfg.DB.GetAllActiveProducts(ctx)
	}

	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_products",
			"query failed",
			"Failed to fetch products",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't fetch products")
		return
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_products", "Get all products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, products)
}

func (apicfg *HandlersProductConfig) HandlerGetProductByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		apicfg.LogHandlerError(
			ctx,
			"get_product_by_id",
			"missing product id",
			"ID of product is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	var (
		product database.Product
		err     error
	)

	if user.Role == "admin" {
		product, err = apicfg.DB.GetProductByID(ctx, productID)
	} else {
		product, err = apicfg.DB.GetActiveProductByID(ctx, productID)
	}

	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"get_product_by_id",
			"query failed",
			"Product not found or error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get_product_by_id", "Get products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, product)
}
