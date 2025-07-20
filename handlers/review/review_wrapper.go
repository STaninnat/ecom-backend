package reviewhandlers

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
)

// ReviewService defines the business logic interface for reviews.
// Provides methods for creating, reading, updating, and deleting reviews,
// as well as paginated retrieval with filtering and sorting capabilities.
//
//go:generate mockery --name=ReviewService --output=./mocks --case=underscore
type ReviewService interface {
	CreateReview(ctx context.Context, review *models.Review) error
	GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error)
	GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error)
	GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error)
	UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error
	DeleteReviewByID(ctx context.Context, reviewID string) error
	GetReviewsByProductIDPaginated(ctx context.Context, productID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error)
	GetReviewsByUserIDPaginated(ctx context.Context, userID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error)
}

// HandlersReviewConfig contains configuration and dependencies for review handlers.
// Embeds HandlersConfig, provides logger, reviewService, and thread safety.
// Manages the lifecycle of review service instances with proper synchronization.
type HandlersReviewConfig struct {
	*handlers.HandlersConfig
	Logger        handlers.HandlerLogger
	ReviewService ReviewService
	ReviewMutex   sync.RWMutex
}

// InitReviewService initializes the review service with the current configuration.
// Validates that the handlers config is initialized and sets up the logger if not already configured.
// Thread-safe operation using mutex for concurrent access.
// Parameters:
//   - service: ReviewService instance to be initialized
//
// Returns:
//   - error: nil on success, error if handlers config is not initialized
func (cfg *HandlersReviewConfig) InitReviewService(service ReviewService) error {
	if cfg.HandlersConfig == nil {
		return errors.New("handlers config not initialized")
	}
	cfg.ReviewMutex.Lock()
	defer cfg.ReviewMutex.Unlock()
	cfg.ReviewService = service
	if cfg.Logger == nil {
		cfg.Logger = cfg.HandlersConfig // HandlersConfig implements HandlerLogger
	}
	return nil
}

// GetReviewService returns the review service instance (thread-safe).
// Uses read lock for concurrent access to the service instance.
// Returns:
//   - ReviewService: the current review service instance
func (cfg *HandlersReviewConfig) GetReviewService() ReviewService {
	cfg.ReviewMutex.RLock()
	service := cfg.ReviewService
	cfg.ReviewMutex.RUnlock()
	return service
}

// handleReviewError maps service errors to HTTP responses and logs them.
// Converts AppError types to appropriate HTTP status codes and logs the error details.
// Parameters:
//   - w: http.ResponseWriter for sending the error response
//   - r: *http.Request containing the request context
//   - err: error that occurred during operation
//   - operation: string identifier for the operation that failed
//   - ip: string representing the client IP address
//   - userAgent: string representing the client user agent
func (cfg *HandlersReviewConfig) handleReviewError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	ctx := r.Context()

	if appErr, ok := err.(*handlers.AppError); ok {
		switch appErr.Code {
		case "not_found":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusNotFound, appErr.Message)
		case "unauthorized":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusForbidden, appErr.Message)
		case "invalid_request":
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusBadRequest, appErr.Message)
		default:
			cfg.Logger.LogHandlerError(ctx, operation, appErr.Code, appErr.Message, ip, userAgent, appErr.Err)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	} else {
		cfg.Logger.LogHandlerError(ctx, operation, "internal_error", "Unknown error occurred", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// ReviewCreateRequest is the DTO for creating a review.
// Contains all required fields for review creation with validation rules.
// Validation: Rating 1-5, Comment required, ProductID required
type ReviewCreateRequest struct {
	ProductID string   `json:"product_id"`
	Rating    int      `json:"rating"`
	Comment   string   `json:"comment"`
	MediaURLs []string `json:"media_urls,omitempty"`
}

// PaginatedReviewsResponse is the response for paginated review lists.
// Provides structured response with pagination metadata and review data.
// Supported query params:
//   - page, pageSize: pagination
//   - rating: exact rating (1-5)
//   - min_rating, max_rating: rating range
//   - from, to: created_at date range (RFC3339)
//   - has_media: true/false (reviews with media)
//   - sort: date_desc, date_asc, rating_desc, rating_asc, updated_desc, updated_asc, comment_length_desc, comment_length_asc
type PaginatedReviewsResponse struct {
	Data       any    `json:"data"`
	TotalCount int64  `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
	HasNext    bool   `json:"hasNext"`
	HasPrev    bool   `json:"hasPrev"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
}

// ReviewUpdateRequest is the DTO for updating a review.
// Contains fields that can be updated for an existing review.
// Validation: Rating 1-5, Comment required
type ReviewUpdateRequest struct {
	Rating    int      `json:"rating"`
	Comment   string   `json:"comment"`
	MediaURLs []string `json:"media_urls,omitempty"`
}
