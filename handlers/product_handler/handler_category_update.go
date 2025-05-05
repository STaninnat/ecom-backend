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
	"github.com/lib/pq"
)

func (apicfg *HandlersProductConfig) HandlerUpdateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	var params CategoryWithIDRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apicfg.LogHandlerError(
			r.Context(),
			"update category",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.ID == "" || params.Name == "" {
		apicfg.LogHandlerError(
			r.Context(),
			"update category",
			"missing category id and name",
			"ID and name of category are empty",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "ID and name are required")
		return
	}

	if len(params.Name) > 100 {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Name too long (max 100 characters)")
		return
	}

	if len(params.Description) > 500 {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Description too long (max 500 characters)")
		return
	}

	err := apicfg.DB.UpdateCategories(r.Context(), database.UpdateCategoriesParams{
		ID:          params.ID,
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			apicfg.LogHandlerError(
				r.Context(),
				"update category",
				"update category failed",
				"Error category name already exists",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusConflict, "Category name already exists")
			return
		}

		apicfg.LogHandlerError(
			r.Context(),
			"update category",
			"update category failed",
			"Error updating category",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't update category")
		return
	}

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, user.ID)
	apicfg.LogHandlerSuccess(ctxWithUserID, "update category", "Updated category successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusNoContent, map[string]string{
		"message": "Updated category successful",
	})
}
