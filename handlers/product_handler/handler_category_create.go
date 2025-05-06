package producthandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (apicfg *HandlersProductConfig) HandlerCreateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var params CategoryWithIDRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			ctx,
			"create_category",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Name == "" {
		apicfg.LogHandlerError(
			ctx,
			"create_category",
			"missing category name",
			"Name of category is empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// This part can be implemented on the frontend instead.
	if len(params.Name) > 100 {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Name too long (max 100 characters)")
		return
	}

	if len(params.Description) > 500 {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Description too long (max 500 characters)")
		return
	}
	// -----------------

	timeNow := time.Now().UTC()

	err := apicfg.DB.CreateCategory(ctx, database.CreateCategoryParams{
		ID:          uuid.New().String(),
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			apicfg.LogHandlerError(
				ctx,
				"create_category",
				"create category failed",
				"Error category name already exists",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusConflict, "Category name already exists")
			return
		}

		apicfg.LogHandlerError(
			ctx,
			"create_category",
			"create category failed",
			"Error creating category",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "create_category", "Created category successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Created category successful",
	})
}
