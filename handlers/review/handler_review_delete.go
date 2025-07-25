// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_review_delete.go: Handles review deletion by ID with validation, ownership check, and error handling.

// HandlerDeleteReviewByID handles HTTP DELETE requests to delete a review by its ID.
// Validates the review ID parameter, checks if the review exists, verifies user ownership,
// and delegates deletion to the review service. On success, logs the event and responds
// with a success message; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data with review ID in URL parameters
//   - user: database.User representing the authenticated user
func (cfg *HandlersReviewConfig) HandlerDeleteReviewByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		cfg.Logger.LogHandlerError(ctx, "delete_review_by_id", "invalid_request", "Review ID is required", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Review ID is required")
		return
	}

	review, err := cfg.GetReviewService().GetReviewByID(ctx, reviewID)
	if err != nil {
		cfg.handleReviewError(w, r, err, "delete_review_by_id", ip, userAgent)
		return
	}

	if review.UserID != user.ID {
		cfg.Logger.LogHandlerError(ctx, "delete_review_by_id", "unauthorized", "You can only delete your own reviews", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusForbidden, "You can only delete your own reviews")
		return
	}

	if err := cfg.GetReviewService().DeleteReviewByID(ctx, reviewID); err != nil {
		cfg.handleReviewError(w, r, err, "delete_review_by_id", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "delete_review_by_id", "Review deleted successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.APIResponse{
		Message: "Review deleted successfully",
		Code:    "success",
	})
}
