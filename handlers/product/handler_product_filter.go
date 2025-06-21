package producthandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersProductConfig) HandlerFilterProducts(w http.ResponseWriter, r *http.Request, user *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params FilterProductsRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"filter_products",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	products, err := apicfg.DB.FilterProducts(ctx, database.FilterProductsParams{
		CategoryID: params.CategoryID.NullString,
		IsActive:   params.IsActive.NullBool,
		MinPrice: sql.NullString{
			String: fmt.Sprintf("%f", params.MinPrice.Float64),
			Valid:  params.MinPrice.Valid,
		},
		MaxPrice: sql.NullString{
			String: fmt.Sprintf("%f", params.MaxPrice.Float64),
			Valid:  params.MaxPrice.Valid,
		},
	})
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"filter_products",
			"failed to fetch products",
			"Error filtering products",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't fetch products")
		return
	}

	productResp := struct {
		Products []database.Product `json:"products"`
	}{
		Products: products,
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "filter_products", "Filter products success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, productResp)
}
