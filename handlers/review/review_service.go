package reviewhandlers

import (
	"context"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
)

// Add this interface above reviewServiceImpl
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

// reviewServiceImpl implements ReviewService for business logic
// All errors returned are *handlers.AppError with standardized codes/messages
type reviewServiceImpl struct {
	reviewMongo ReviewMongoAPI
}

// NewReviewService creates a new ReviewService instance
func NewReviewService(reviewMongo ReviewMongoAPI) ReviewService {
	return &reviewServiceImpl{reviewMongo: reviewMongo}
}

// CreateReview creates a new review
func (s *reviewServiceImpl) CreateReview(ctx context.Context, review *models.Review) error {
	if err := s.reviewMongo.CreateReview(ctx, review); err != nil {
		return &handlers.AppError{Code: "create_failed", Message: "Failed to create review", Err: err}
	}
	return nil
}

// GetReviewByID fetches a review by its ID
func (s *reviewServiceImpl) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	review, err := s.reviewMongo.GetReviewByID(ctx, reviewID)
	if err != nil {
		if err.Error() == "review not found" {
			return nil, &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get review", Err: err}
	}
	return review, nil
}

// GetReviewsByProductID fetches all reviews for a product
func (s *reviewServiceImpl) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	reviews, err := s.reviewMongo.GetReviewsByProductID(ctx, productID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by product", Err: err}
	}
	return reviews, nil
}

// GetReviewsByUserID fetches all reviews by a user
func (s *reviewServiceImpl) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	reviews, err := s.reviewMongo.GetReviewsByUserID(ctx, userID)
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by user", Err: err}
	}
	return reviews, nil
}

// UpdateReviewByID updates a review by its ID
func (s *reviewServiceImpl) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	if err := s.reviewMongo.UpdateReviewByID(ctx, reviewID, updatedReview); err != nil {
		if err.Error() == "review not found" {
			return &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return &handlers.AppError{Code: "update_failed", Message: "Failed to update review", Err: err}
	}
	return nil
}

// DeleteReviewByID deletes a review by its ID
func (s *reviewServiceImpl) DeleteReviewByID(ctx context.Context, reviewID string) error {
	if err := s.reviewMongo.DeleteReviewByID(ctx, reviewID); err != nil {
		if err.Error() == "review not found" {
			return &handlers.AppError{Code: "not_found", Message: "Review not found", Err: err}
		}
		return &handlers.AppError{Code: "delete_failed", Message: "Failed to delete review", Err: err}
	}
	return nil
}

// GetReviewsByProductIDPaginated fetches paginated, filtered, and sorted reviews for a product
func (s *reviewServiceImpl) GetReviewsByProductIDPaginated(ctx context.Context, productID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	filter := map[string]any{"product_id": productID}
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
	findSort := parseSortOption(sort)
	result, err := s.reviewMongo.GetReviewsByProductIDPaginated(ctx, productID, &intmongo.PaginationOptions{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Sort:     findSort,
		Filter:   filter,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by product (paginated)", Err: err}
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

// GetReviewsByUserIDPaginated fetches paginated, filtered, and sorted reviews for a user
func (s *reviewServiceImpl) GetReviewsByUserIDPaginated(ctx context.Context, userID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	filter := map[string]any{"user_id": userID}
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
	findSort := parseSortOption(sort)
	result, err := s.reviewMongo.GetReviewsByUserIDPaginated(ctx, userID, &intmongo.PaginationOptions{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Sort:     findSort,
		Filter:   filter,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "get_failed", Message: "Failed to get reviews by user (paginated)", Err: err}
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

// parseSortOption converts a sort string to a mongo sort option
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
