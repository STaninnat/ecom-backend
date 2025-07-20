package models

import (
	"time"
)

// Review represents a product review submitted by a user.
// It contains rating, comment, and optional media attachments.
type Review struct {
	ID        string    `bson:"_id,omitempty" json:"id"`                          // Unique identifier for the review
	UserID    string    `bson:"user_id" json:"user_id"`                           // ID of the user who wrote the review
	ProductID string    `bson:"product_id" json:"product_id"`                     // ID of the product being reviewed
	Rating    int       `bson:"rating" json:"rating"`                             // Numeric rating (typically 1-5 stars)
	Comment   string    `bson:"comment,omitempty" json:"comment,omitempty"`       // Optional text review
	MediaURLs []string  `bson:"media_urls,omitempty" json:"media_urls,omitempty"` // Optional image/video URLs
	CreatedAt time.Time `bson:"created_at" json:"created_at"`                     // When the review was created
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`                     // When the review was last updated
}
