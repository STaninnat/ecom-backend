package intmongo

import (
	"context"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// --- ReviewMongo ---
type ReviewMongo struct {
	Collection CollectionInterface
}

func NewReviewMongo(db *mongo.Database) *ReviewMongo {
	return &ReviewMongo{
		Collection: &MongoCollectionAdapter{
			Inner: db.Collection("reviews"),
		},
	}
}

func (r *ReviewMongo) CreateReview(ctx context.Context, review *models.Review) error {
	timeNow := time.Now().UTC()
	review.CreatedAt = timeNow
	review.UpdatedAt = timeNow

	if review.ID == "" {
		review.ID = bson.NewObjectID().Hex()
	}

	_, err := r.Collection.InsertOne(ctx, review)
	return err
}

func (r *ReviewMongo) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	filter := bson.M{"product_id": productID}

	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	cursor := NewCursorAdapter(cur)
	defer cursor.Close(ctx)

	var reviews []*models.Review
	for cursor.Next(ctx) {
		var review models.Review
		if err := cursor.Decode(&review); err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (r *ReviewMongo) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	filter := bson.M{"user_id": userID}

	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	cursor := NewCursorAdapter(cur)
	defer cursor.Close(ctx)

	var reviews []*models.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *ReviewMongo) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	filter := bson.M{"_id": reviewID}

	res := r.Collection.FindOne(ctx, filter)
	result := NewSingleResultAdapter(res)

	var review models.Review
	err := result.Decode(&review)
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *ReviewMongo) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	update := bson.M{
		"$set": bson.M{
			"rating":     updatedReview.Rating,
			"comment":    updatedReview.Comment,
			"media_urls": updatedReview.MediaURLs,
			"updated_at": time.Now(),
		},
	}

	filter := bson.M{"_id": reviewID}
	_, err := r.Collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *ReviewMongo) DeleteReviewByID(ctx context.Context, reviewID string) error {
	filter := bson.M{"_id": reviewID}
	_, err := r.Collection.DeleteOne(ctx, filter)
	return err
}
