package repository

import (
	"context"
	"fmt"

	"github.com/atavada/project-management-saas/internal/database"
	"github.com/atavada/project-management-saas/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (clerk_user_id, email, first_name, last_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, clerk_user_id, email, first_name, last_name, avatar_url, created_at, updated_at
	`
	
	var result models.User
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		user.ClerkUserID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
	).Scan(
		&result.ID,
		&result.ClerkUserID,
		&result.Email,
		&result.FirstName,
		&result.LastName,
		&result.AvatarURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &result, nil
}

func (r *UserRepository) GetByClerkID(ctx context.Context, clerkUserID string) (*models.User, error) {
	query := `
		SELECT id, clerk_user_id, email, first_name, last_name, avatar_url, created_at, updated_at
		FROM users
		WHERE clerk_user_id = $1
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, clerkUserID).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
} 

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, clerk_user_id, email, first_name, last_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (clerk_user_id, email, first_name, last_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (clerk_user_id)
		DO UPDATE SET
				email = EXCLUDED.email,
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				avatar_url = EXCLUDED.avatar_url,
				updated_at = CURRENT_TIMESTAMP
		RETURNING id, clerk_user_id, email, first_name, last_name, avatar_url, created_at, updated_at
	`

	var result models.User
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		user.ClerkUserID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
	).Scan(
		&result.ID,
		&result.ClerkUserID,
		&result.Email,
		&result.FirstName,
		&result.LastName,
		&result.AvatarURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error upserting user: %w", err)
	}

	return &result, nil
}

