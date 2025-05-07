package models

import (
	"time"
)

type Review struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	ProductID string    `bson:"product_id" json:"product_id"`
	Rating    int       `bson:"rating" json:"rating"`
	Comment   string    `bson:"comment,omitempty" json:"comment,omitempty"`
	MediaURLs []string  `bson:"media_urls,omitempty" json:"media_urls,omitempty"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
