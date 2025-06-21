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

func (apicfg *HandlersReviewConfig) HandlerGetReviewsByProductID(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "product_id")
	if productID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_review_by_id",
			"missing product id",
			"Product ID not found in URL",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	reviews, err := apicfg.ReviewMG.GetReviewsByProductID(r.Context(), productID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_review_by_id",
			"get review failed",
			"Error getting review",
			ip, userAgent, nil,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to get reviews")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "create_review", "Got review successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, reviews)
}

func (apicfg *HandlersReviewConfig) HandlerGetReviewsByUserID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviews, err := apicfg.ReviewMG.GetReviewsByUserID(ctx, user.ID)
	if err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_reviews_by_user",
			"failed to get user reviews",
			"Error retrieving reviews by user",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to get user reviews")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "get_reviews_by_user", "Got user reviews successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, reviews)
}

func (apicfg *HandlersReviewConfig) HandlerGetReviewByID(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"get_review_by_id",
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
			"get_review_by_id",
			"review not found",
			"Review lookup failed",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusNotFound, "Review not found")
		return
	}

	apicfg.HandlersConfig.LogHandlerSuccess(ctx, "get_review_by_id", "Got review successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, review)
}
