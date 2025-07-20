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

// HandlerGetAllProducts handles HTTP GET requests to retrieve all products.
// Checks if the user is an admin, fetches products accordingly, logs the event, and responds with the product list.
// On error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: *database.User representing the authenticated user (may be nil)
func (cfg *HandlersProductConfig) HandlerGetAllProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	isAdmin := user != nil && user.Role == "admin"
	products, err := cfg.GetProductService().GetAllProducts(ctx, isAdmin)
	if err != nil {
		cfg.handleProductError(w, r, err, "get_products", ip, userAgent)
		return
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_products", "Get all products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, products)
}

// HandlerGetProductByID handles HTTP GET requests to retrieve a product by its ID.
// Extracts the product ID from the URL, checks admin status, fetches the product, logs the event, and responds with the product data.
// On error or missing ID, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersProductConfig) HandlerGetProductByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		cfg.Logger.LogHandlerError(
			ctx,
			"get_product_by_id",
			"invalid_request",
			"Missing product ID",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	isAdmin := user.Role == "admin"
	product, err := cfg.GetProductService().GetProductByID(ctx, productID, isAdmin)
	if err != nil {
		cfg.handleProductError(w, r, err, "get_product_by_id", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_product_by_id", "Get products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, product)
}
