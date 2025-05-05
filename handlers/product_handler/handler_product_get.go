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

	var (
		products []database.Product
		err      error
	)

	if user.Role == "admin" {
		products, err = apicfg.DB.GetAllProducts(r.Context())
	} else {
		products, err = apicfg.DB.GetAllActiveProducts(r.Context())
	}

	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"get products",
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

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get products", "Get all products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, products)
}

func (apicfg *HandlersProductConfig) HandlerGetProductByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	productID := chi.URLParam(r, "id")
	if productID == "" {
		apicfg.LogHandlerError(
			r.Context(),
			"get product by id",
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
		product, err = apicfg.DB.GetProductByID(r.Context(), productID)
	} else {
		product, err = apicfg.DB.GetActiveProductByID(r.Context(), productID)
	}

	if err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"get product by id",
			"query failed",
			"Product not found or error",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "get product by id", "Get products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, product)
}
