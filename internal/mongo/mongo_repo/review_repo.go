package mongorepo

import (
	"context"

	"github.com/STaninnat/ecom-backend/models"
)

type ReviewRepository interface {
	CreateReview(ctx context.Context, review *models.Review) error
	GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error)
	GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error)
	GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error)
	UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error
	DeleteReviewByID(ctx context.Context, reviewID string) error
}
