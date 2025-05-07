package reviewhandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersReviewConfig) HandlerUpdateReviewByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_review_by_id",
			"missing review id",
			"Review ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Review ID is required")
		return
	}

	var input models.Review
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_review_by_id",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.UserID != user.ID {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_review_by_id",
			"unauthorized",
			"User does not own the review",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusForbidden, "You can only update your own reviews")
		return
	}

	if err := apicfg.ReviewRepo.UpdateReviewByID(ctx, reviewID, &input); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"update_review_by_id",
			"update review failed",
			"Error updating review",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to update review")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "update_review_by_id", "Review updated successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, input)
}
