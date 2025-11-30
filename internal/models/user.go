package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
		ID           uuid.UUID  `json:"id"`
    ClerkUserID  string     `json:"clerk_user_id"`
    Email        string     `json:"email"`
    FirstName    string     `json:"first_name"`
    LastName     string     `json:"last_name"`
    AvatarURL    string     `json:"avatar_url"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateUserRequest struct {
		ClerkUserID string `json:"clerk_user_id" validate:"required"`
    Email       string `json:"email" validate:"required,email"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    AvatarURL   string `json:"avatar_url"`
}