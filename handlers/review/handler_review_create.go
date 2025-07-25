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
)

// handler_review_create.go: Validates and handles review creation requests with logging and error handling.

// Validate checks the request for required fields and valid values.
// Ensures that the ProductID is provided, the Rating is within the valid range (1-5), and the Comment field is not empty.
// Returns an AppError if any validation fails.
// Returns:
//   - error: AppError with appropriate code and message if validation fails, nil otherwise
func (r *ReviewCreateRequest) Validate() error {
	if r.ProductID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	if r.Rating < 1 || r.Rating > 5 {
		return &handlers.AppError{Code: "invalid_request", Message: "Rating must be between 1 and 5"}
	}
	if r.Comment == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Comment is required"}
	}
	return nil
}

// HandlerCreateReview handles HTTP POST requests to create a new review.
// @Summary      Create review
// @Description  Creates a new review for a product
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        review  body  object{}  true  "Review payload"
// @Success      201  {object}  handlers.APIResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/reviews/ [post]
func (cfg *HandlersReviewConfig) HandlerCreateReview(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req ReviewCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"create_review",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		cfg.handleReviewError(w, r, err, "create_review", ip, userAgent)
		return
	}

	review := &models.Review{
		UserID:    user.ID,
		ProductID: req.ProductID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		MediaURLs: req.MediaURLs,
	}

	if err := cfg.GetReviewService().CreateReview(ctx, review); err != nil {
		cfg.handleReviewError(w, r, err, "create_review", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "create_review", "Review created successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.APIResponse{
		Message: "Review created successfully",
		Code:    "success",
		Data:    review,
	})
}
