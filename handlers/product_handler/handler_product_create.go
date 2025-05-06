package producthandlers

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
	"github.com/lib/pq"
)

func (apicfg *HandlersProductConfig) HandlerCreateProduct(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params ProductRequest

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_product",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.CategoryID == "" || params.Name == "" || params.Price <= 0 || params.Stock < 0 {
		apicfg.LogHandlerError(
			ctx,
			"create_product",
			"missing fields",
			"Required fields are missing",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields")
		return
	}

	id := uuid.New().String()
	timeNow := time.Now().UTC()
	isActive := true
	if params.IsActive != nil {
		isActive = *params.IsActive
	}

	err := apicfg.DB.CreateProduct(ctx, database.CreateProductParams{
		ID:          id,
		CategoryID:  utils.ToNullString(params.CategoryID),
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		Price:       fmt.Sprintf("%.2f", params.Price),
		Stock:       params.Stock,
		ImageUrl:    utils.ToNullString(params.ImageURL),
		IsActive:    isActive,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			apicfg.LogHandlerError(
				ctx,
				"create_product",
				"create product failed",
				"Error product name already exists",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusConflict, "Category name already exists")
			return
		}

		apicfg.LogHandlerError(
			ctx,
			"create_product",
			"create product failed",
			"Error creating product",
			ip, userAgent, err,
		)

		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't create product")
		return
	}

	productResp := map[string]string{
		"message":    "Product created successfully",
		"product_id": id,
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "create_product", "Created product successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, productResp)
}
