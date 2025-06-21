package reviewhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/go-chi/chi/v5"
)

func (apicfg *HandlersReviewConfig) HandlerDeleteReviewByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"delete_review_by_id",
			"missing review id",
			"Review ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Review ID is required")
		return
	}

	review, err := apicfg.ReviewMG.GetReviewByID(ctx, reviewID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"delete_review_by_id",
			"get review failed",
			"Error getting review",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to get review")
		return
	}

	if review.UserID != user.ID {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"delete_review_by_id",
			"unauthorized",
			"User does not own the review",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusForbidden, "You can only delete your own reviews")
		return
	}

	if err := apicfg.ReviewMG.DeleteReviewByID(ctx, reviewID); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"delete_review_by_id",
			"delete review failed",
			"Error deleting review",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to delete review")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "delete_review_by_id", "Review deleted successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Review deleted successfully",
	})
}
