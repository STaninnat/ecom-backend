// Package mongo provides MongoDB repositories and helpers for the ecom-backend project.
package intmongo

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/STaninnat/ecom-backend/models"
)

// review.go: MongoDB repository and operations for product reviews.

// ReviewMongo handles review operations in MongoDB.
type ReviewMongo struct {
	Collection CollectionInterface
}

// NewReviewMongo creates a new ReviewMongo instance for the given MongoDB database.
func NewReviewMongo(db *mongo.Database) *ReviewMongo {
	return &ReviewMongo{
		Collection: &MongoCollectionAdapter{
			Inner: db.Collection("reviews"),
		},
	}
}

// CreateReview creates a new review in the database.
func (r *ReviewMongo) CreateReview(ctx context.Context, review *models.Review) error {
	if review == nil {
		return fmt.Errorf("review cannot be nil")
	}

	timeNow := time.Now().UTC()
	review.CreatedAt = timeNow
	review.UpdatedAt = timeNow

	if review.ID == "" {
		review.ID = bson.NewObjectID().Hex()
	}

	_, err := r.Collection.InsertOne(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// CreateReviews creates multiple reviews in a single operation.
func (r *ReviewMongo) CreateReviews(ctx context.Context, reviews []*models.Review) error {
	if len(reviews) == 0 {
		return fmt.Errorf("reviews slice cannot be empty")
	}

	timeNow := time.Now().UTC()
	documents := make([]any, len(reviews))

	for i, review := range reviews {
		if review == nil {
			return fmt.Errorf("review at index %d cannot be nil", i)
		}

		review.CreatedAt = timeNow
		review.UpdatedAt = timeNow

		if review.ID == "" {
			review.ID = bson.NewObjectID().Hex()
		}

		documents[i] = review
	}

	_, err := r.Collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to create reviews: %w", err)
	}

	return nil
}

// getReviewsByField is a shared helper for retrieving reviews by a specific field (e.g., product_id, user_id).
func (r *ReviewMongo) getReviewsByField(ctx context.Context, filterKey, displayName, value string) ([]*models.Review, error) {
	if value == "" {
		return nil, fmt.Errorf("%s cannot be empty", displayName)
	}
	filter := bson.M{filterKey: value}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by %s: %w", displayName, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			fmt.Printf("cursor.Close failed: %v\n", err)
		}
	}()
	var reviews []*models.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}
	return reviews, nil
}

// getReviewsByFieldPaginated is a shared helper for retrieving paginated reviews by a specific field.
func (r *ReviewMongo) getReviewsByFieldPaginated(ctx context.Context, filterKey, displayName, value string, pagination *PaginationOptions) (*PaginatedResult[*models.Review], error) {
	if value == "" {
		return nil, fmt.Errorf("%s cannot be empty", displayName)
	}
	if pagination == nil {
		pagination = NewPaginationOptions(1, 10)
	}
	filter := bson.M{filterKey: value}
	if pagination.Filter != nil {
		maps.Copy(filter, pagination.Filter)
	}
	// Get total count
	totalCount, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count reviews: %w", err)
	}
	// Calculate pagination
	skip := (pagination.Page - 1) * pagination.PageSize
	findOptions := options.Find().
		SetLimit(pagination.PageSize).
		SetSkip(skip).
		SetSort(pagination.Sort)
	cursor, err := r.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by %s: %w", displayName, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			fmt.Printf("cursor.Close failed: %v\n", err)
		}
	}()
	var reviews []*models.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}
	// Calculate pagination metadata
	totalPages := (totalCount + pagination.PageSize - 1) / pagination.PageSize
	hasNext := pagination.Page < totalPages
	hasPrev := pagination.Page > 1
	return &PaginatedResult[*models.Review]{
		Data:       reviews,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// Refactored public methods to use the shared helpers
func (r *ReviewMongo) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	return r.getReviewsByField(ctx, "product_id", "product ID", productID)
}

func (r *ReviewMongo) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	return r.getReviewsByField(ctx, "user_id", "user ID", userID)
}

func (r *ReviewMongo) GetReviewsByProductIDPaginated(ctx context.Context, productID string, pagination *PaginationOptions) (*PaginatedResult[*models.Review], error) {
	return r.getReviewsByFieldPaginated(ctx, "product_id", "product ID", productID, pagination)
}

func (r *ReviewMongo) GetReviewsByUserIDPaginated(ctx context.Context, userID string, pagination *PaginationOptions) (*PaginatedResult[*models.Review], error) {
	return r.getReviewsByFieldPaginated(ctx, "user_id", "user ID", userID, pagination)
}

// GetReviewByID retrieves a specific review by its ID.
func (r *ReviewMongo) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	if reviewID == "" {
		return nil, fmt.Errorf("review ID cannot be empty")
	}

	filter := bson.M{"_id": reviewID}

	result := r.Collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("review not found")
		}
		return nil, fmt.Errorf("failed to find review: %w", result.Err())
	}

	var review models.Review
	if err := result.Decode(&review); err != nil {
		return nil, fmt.Errorf("failed to decode review: %w", err)
	}

	return &review, nil
}

// UpdateReviewByID updates an existing review by its ID.
func (r *ReviewMongo) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	if reviewID == "" {
		return fmt.Errorf("review ID cannot be empty")
	}
	if updatedReview == nil {
		return fmt.Errorf("updated review cannot be nil")
	}

	filter := bson.M{"_id": reviewID}
	update := bson.M{
		"$set": bson.M{
			"rating":     updatedReview.Rating,
			"comment":    updatedReview.Comment,
			"media_urls": updatedReview.MediaURLs,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// UpdateReviewsByProductID updates all reviews for a specific product.
func (r *ReviewMongo) UpdateReviewsByProductID(ctx context.Context, productID string, update bson.M) error {
	if productID == "" {
		return fmt.Errorf("product ID cannot be empty")
	}

	filter := bson.M{"product_id": productID}
	update["$set"] = bson.M{
		"updated_at": time.Now().UTC(),
	}

	_, err := r.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update reviews: %w", err)
	}

	return nil
}

// DeleteReviewByID deletes a review by its ID.
func (r *ReviewMongo) DeleteReviewByID(ctx context.Context, reviewID string) error {
	if reviewID == "" {
		return fmt.Errorf("review ID cannot be empty")
	}

	filter := bson.M{"_id": reviewID}

	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// DeleteReviewsByUserID deletes all reviews by a specific user.
func (r *ReviewMongo) DeleteReviewsByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	filter := bson.M{"user_id": userID}

	_, err := r.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete reviews: %w", err)
	}

	return nil
}

// GetProductRatingStats gets rating statistics for a product.
func (r *ReviewMongo) GetProductRatingStats(ctx context.Context, productID string) (map[string]any, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	pipeline := []bson.M{
		{"$match": bson.M{"product_id": productID}},
		{"$group": bson.M{
			"_id":           nil,
			"averageRating": bson.M{"$avg": "$rating"},
			"totalReviews":  bson.M{"$sum": 1},
			"ratingCounts":  bson.M{"$push": "$rating"},
		}},
		{"$project": bson.M{
			"_id":           0,
			"averageRating": bson.M{"$round": []any{"$averageRating", 2}},
			"totalReviews":  1,
			"ratingCounts":  1,
		}},
	}

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate rating stats: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			fmt.Printf("cursor.Close failed: %v\n", err)
		}
	}()

	var results []map[string]any
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	if len(results) == 0 {
		return map[string]any{
			"averageRating": 0.0,
			"totalReviews":  0,
			"ratingCounts":  []int{},
		}, nil
	}

	return results[0], nil
}
