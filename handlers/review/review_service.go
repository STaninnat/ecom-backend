// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
)

// review_service.go: Implements review business logic with error handling, using MongoDB for data access and filtering.

const (
	reviewNotFoundMsg = "review not found"
)

// ReviewMongoAPI defines the interface for MongoDB operations on reviews.
// Provides data access layer methods for review CRUD operations and paginated queries.
type ReviewMongoAPI interface {
	CreateReview(ctx context.Context, review *models.Review) error
	GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error)
	GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error)
	GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error)
	UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error
	DeleteReviewByID(ctx context.Context, reviewID string) error
	GetReviewsByProductIDPaginated(ctx context.Context, productID string, opts *intmongo.PaginationOptions) (*intmongo.PaginatedResult[*models.Review], error)
	GetReviewsByUserIDPaginated(ctx context.Context, userID string, opts *intmongo.PaginationOptions) (*intmongo.PaginatedResult[*models.Review], error)
}

// reviewServiceImpl implements ReviewService for business logic.
// All errors returned are *handlers.AppError with standardized codes/messages.
// Provides business logic layer between handlers and data access layer.
type reviewServiceImpl struct {
	reviewMongo ReviewMongoAPI
}

// NewReviewService creates a new ReviewService instance.
// Initializes the review service with the provided MongoDB API implementation.
// Parameters:
//   - reviewMongo: ReviewMongoAPI implementation for data access
//
// Returns:
//   - ReviewService: configured review service instance
func NewReviewService(reviewMongo ReviewMongoAPI) ReviewService {
	return &reviewServiceImpl{reviewMongo: reviewMongo}
}

// CreateReview creates a new review.
// Delegates to the MongoDB API and wraps any errors in standardized AppError format.
// Parameters:
//   - ctx: context.Context for the operation
//   - review: *models.Review to be created
//
// Returns:
//   - error: nil on success, AppError with "create_failed" code on failure
func (s *reviewServiceImpl) CreateReview(ctx context.Context, review *models.Review) error {
	if err := s.reviewMongo.CreateReview(ctx, review); err != nil {
		return &handlers.AppError{Code: "create_failed", Message: "Failed to create review", Err: err}
	}
	return nil
}

// GetReviewByID fetches a review by its ID.
// Delegates to the MongoDB API and handles "not found" cases with appropriate error codes.
// Parameters:
//   - ctx: context.Context for the operation
//   - reviewID: string identifier of the review to fetch
//
// Returns:
//   - *models.Review: the found review, nil if not found
//   - error: nil on success, AppError with "not_found" or "get_failed" code on failure
func (s *reviewServiceImpl) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	review, err := s.reviewMongo.GetReviewByID(ctx, reviewID)
	if err != nil {
		if err.Error() == reviewNotFoundMsg {
			return nil, &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get review", Err: err}
	}
	return review, nil
}

// GetReviewsByProductID fetches all reviews for a product.
// Delegates to the MongoDB API and wraps any errors in standardized AppError format.
// Parameters:
//   - ctx: context.Context for the operation
//   - productID: string identifier of the product
//
// Returns:
//   - []*models.Review: list of reviews for the product
//   - error: nil on success, AppError with "get_failed" code on failure
func (s *reviewServiceImpl) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	reviews, err := s.reviewMongo.GetReviewsByProductID(ctx, productID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by product", Err: err}
	}
	return reviews, nil
}

// GetReviewsByUserID fetches all reviews by a user.
// Delegates to the MongoDB API and wraps any errors in standardized AppError format.
// Parameters:
//   - ctx: context.Context for the operation
//   - userID: string identifier of the user
//
// Returns:
//   - []*models.Review: list of reviews by the user
//   - error: nil on success, AppError with "get_failed" code on failure
func (s *reviewServiceImpl) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	reviews, err := s.reviewMongo.GetReviewsByUserID(ctx, userID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by user", Err: err}
	}
	return reviews, nil
}

// UpdateReviewByID updates a review by its ID.
// Delegates to the MongoDB API and handles "not found" cases with appropriate error codes.
// Parameters:
//   - ctx: context.Context for the operation
//   - reviewID: string identifier of the review to update
//   - updatedReview: *models.Review containing the updated data
//
// Returns:
//   - error: nil on success, AppError with "not_found" or "update_failed" code on failure
func (s *reviewServiceImpl) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	if err := s.reviewMongo.UpdateReviewByID(ctx, reviewID, updatedReview); err != nil {
		if err.Error() == reviewNotFoundMsg {
			return &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return &handlers.AppError{Code: "update_failed", Message: "Failed to update review", Err: err}
	}
	return nil
}

// DeleteReviewByID deletes a review by its ID.
// Delegates to the MongoDB API and handles "not found" cases with appropriate error codes.
// Parameters:
//   - ctx: context.Context for the operation
//   - reviewID: string identifier of the review to delete
//
// Returns:
//   - error: nil on success, AppError with "not_found" or "delete_failed" code on failure
func (s *reviewServiceImpl) DeleteReviewByID(ctx context.Context, reviewID string) error {
	if err := s.reviewMongo.DeleteReviewByID(ctx, reviewID); err != nil {
		if err.Error() == reviewNotFoundMsg {
			return &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return &handlers.AppError{Code: "delete_failed", Message: "Failed to delete review", Err: err}
	}
	return nil
}

// buildReviewFilter constructs a MongoDB filter for reviews based on the root key and common filter parameters.
func buildReviewFilter(rootKey, rootValue string, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool) map[string]any {
	filter := map[string]any{rootKey: rootValue}
	if rating != nil {
		filter["rating"] = *rating
	}
	if minRating != nil || maxRating != nil {
		ratingRange := map[string]any{}
		if minRating != nil {
			ratingRange["$gte"] = *minRating
		}
		if maxRating != nil {
			ratingRange["$lte"] = *maxRating
		}
		filter["rating"] = ratingRange
	}
	if from != nil || to != nil {
		dateRange := map[string]any{}
		if from != nil {
			dateRange["$gte"] = *from
		}
		if to != nil {
			dateRange["$lte"] = *to
		}
		filter["created_at"] = dateRange
	}
	if hasMedia != nil {
		if *hasMedia {
			filter["media_urls.0"] = map[string]any{"$exists": true}
		} else {
			filter["media_urls"] = map[string]any{"$size": 0}
		}
	}
	return filter
}

// getReviewsByFieldPaginated is a shared helper for paginated review retrieval by a field (product_id or user_id).
func (s *reviewServiceImpl) getReviewsByFieldPaginated(
	ctx context.Context,
	rootKey, id string,
	mongoFunc func(context.Context, string, *intmongo.PaginationOptions) (*intmongo.PaginatedResult[*models.Review], error),
	page, pageSize int,
	rating, minRating, maxRating *int,
	from, to *time.Time,
	hasMedia *bool,
	sort, errMsg string,
) (any, error) {
	filter := buildReviewFilter(rootKey, id, rating, minRating, maxRating, from, to, hasMedia)
	findSort := parseSortOption(sort)
	result, err := mongoFunc(ctx, id, &intmongo.PaginationOptions{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Sort:     findSort,
		Filter:   filter,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: errMsg, Err: err}
	}
	return PaginatedReviewsResponse{
		Data:       result.Data,
		TotalCount: result.TotalCount,
		Page:       int(result.Page),
		PageSize:   int(result.PageSize),
		TotalPages: int(result.TotalPages),
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}, nil
}

func (s *reviewServiceImpl) GetReviewsByProductIDPaginated(ctx context.Context, productID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	return s.getReviewsByFieldPaginated(
		ctx,
		"product_id",
		productID,
		s.reviewMongo.GetReviewsByProductIDPaginated,
		page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort,
		"Failed to get reviews by product (paginated)",
	)
}

func (s *reviewServiceImpl) GetReviewsByUserIDPaginated(ctx context.Context, userID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	return s.getReviewsByFieldPaginated(
		ctx,
		"user_id",
		userID,
		s.reviewMongo.GetReviewsByUserIDPaginated,
		page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort,
		"Failed to get reviews by user (paginated)",
	)
}

// parseSortOption converts a sort string to a mongo sort option.
// Maps human-readable sort options to MongoDB sort specifications.
// Supported options: date_desc, date_asc, rating_desc, rating_asc, updated_desc, updated_asc, comment_length_desc, comment_length_asc.
// Parameters:
//   - sort: string representing the sort option
//
// Returns:
//   - map[string]any: MongoDB sort specification, defaults to {"created_at": -1} for unknown options
func parseSortOption(sort string) map[string]any {
	switch sort {
	case "date_desc":
		return map[string]any{"created_at": -1}
	case "date_asc":
		return map[string]any{"created_at": 1}
	case "rating_desc":
		return map[string]any{"rating": -1}
	case "rating_asc":
		return map[string]any{"rating": 1}
	case "updated_desc":
		return map[string]any{"updated_at": -1}
	case "updated_asc":
		return map[string]any{"updated_at": 1}
	case "comment_length_desc":
		return map[string]any{"$expr": map[string]any{"$strLenCP": "$comment"}, "$meta": -1}
	case "comment_length_asc":
		return map[string]any{"$expr": map[string]any{"$strLenCP": "$comment"}, "$meta": 1}
	default:
		return map[string]any{"created_at": -1}
	}
}
