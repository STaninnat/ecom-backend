// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
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

// handler_review_update.go: Handles updating a review by ID: validates input, checks ownership, updates via service, and sends response.

// Validate checks the request for required fields and valid values.
// Ensures that the Rating is within the valid range (1-5) and the Comment field is not empty.
// Returns an AppError if any validation fails.
// Returns:
//   - error: AppError with appropriate code and message if validation fails, nil otherwise
func (r *ReviewUpdateRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return &handlers.AppError{Code: "invalid_request", Message: "Rating must be between 1 and 5"}
	}
	if r.Comment == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Comment is required"}
	}
	return nil
}

// HandlerUpdateReviewByID handles HTTP PUT requests to update an existing review by its ID.
// Validates the review ID parameter, checks if the review exists, verifies user ownership,
// parses and validates the update request, and delegates the update to the review service.
// On success, logs the event and responds with the updated review; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data with review ID in URL parameters and update data in body
//   - user: database.User representing the authenticated user
func (cfg *HandlersReviewConfig) HandlerUpdateReviewByID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		cfg.Logger.LogHandlerError(ctx, "update_review_by_id", "invalid_request", "Review ID is required", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Review ID is required")
		return
	}

	review, err := cfg.GetReviewService().GetReviewByID(ctx, reviewID)
	if err != nil {
		cfg.handleReviewError(w, r, err, "update_review_by_id", ip, userAgent)
		return
	}

	if review.UserID != user.ID {
		cfg.Logger.LogHandlerError(ctx, "update_review_by_id", "unauthorized", "You can only update your own reviews", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusForbidden, "You can only update your own reviews")
		return
	}

	var req ReviewUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"update_review_by_id",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		cfg.handleReviewError(w, r, err, "update_review_by_id", ip, userAgent)
		return
	}

	update := &models.Review{
		UserID:    user.ID,
		ProductID: review.ProductID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		MediaURLs: req.MediaURLs,
	}

	if err := cfg.GetReviewService().UpdateReviewByID(ctx, reviewID, update); err != nil {
		cfg.handleReviewError(w, r, err, "update_review_by_id", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "update_review_by_id", "Review updated successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.APIResponse{
		Message: "Review updated successfully",
		Code:    "success",
		Data:    update,
	})
}
