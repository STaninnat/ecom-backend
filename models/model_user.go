// Package models defines data structures and database models for the ecom-backend project.
package models

import (
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// model_user.go: Defines the User model and mapping from database entities.

// User represents a customer account in the e-commerce system.
// It contains personal information and contact details for order processing.
type User struct {
	ID        string    `json:"id"`         // Unique identifier for the user
	Name      string    `json:"name"`       // Full name of the user
	Email     string    `json:"email"`      // Email address for account access and notifications
	Phone     string    `json:"phone"`      // Phone number for contact purposes
	Address   string    `json:"address"`    // Shipping/billing address
	CreatedAt time.Time `json:"created_at"` // When the user account was created
	UpdatedAt time.Time `json:"updated_at"` // When the user account was last updated
}

// MapUserToResponse converts a database User entity to a User model for API responses.
// It handles nullable fields (Phone and Address) by converting them to empty strings if null.
func MapUserToResponse(user database.User) User {
	return User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: func() string {
			if user.Phone.Valid {
				return user.Phone.String
			}
			return ""
		}(),
		Address: func() string {
			if user.Address.Valid {
				return user.Address.String
			}
			return ""
		}(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
