package models

import (
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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
