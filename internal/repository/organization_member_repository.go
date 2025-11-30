package repository

import (
	"context"
	"fmt"

	"github.com/atavada/project-management-saas/internal/database"
	"github.com/atavada/project-management-saas/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrganizationMemberRepository struct {
    db *database.DB
}

func NewOrganizationMemberRepository(db *database.DB) *OrganizationMemberRepository {
    return &OrganizationMemberRepository{db: db}
}

func (r *OrganizationMemberRepository) Create(ctx context.Context, member *models.OrganizationMember) error {
    query := `
        INSERT INTO organization_members (organization_id, user_id, role, clerk_membership_id)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (organization_id, user_id) DO NOTHING
    `

    _, err := r.db.Pool.Exec(
        ctx,
        query,
        member.OrganizationID,
        member.UserID,
        member.Role,
        member.ClerkMembershipID,
    )

    if err != nil {
        return fmt.Errorf("error creating organization member: %w", err)
    }

    return nil
}

func (r *OrganizationMemberRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
    query := `
        SELECT id, organization_id, user_id, role, clerk_membership_id, joined_at, updated_at
        FROM organization_members
        WHERE organization_id = $1 AND user_id = $2
    `

    var member models.OrganizationMember
    err := r.db.Pool.QueryRow(ctx, query, orgID, userID).Scan(
        &member.ID,
        &member.OrganizationID,
        &member.UserID,
        &member.Role,
        &member.ClerkMembershipID,
        &member.JoinedAt,
        &member.UpdatedAt,
    )

    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("error getting member: %w", err)
    }

    return &member, nil
}

func (r *OrganizationMemberRepository) Delete(ctx context.Context, orgID, userID uuid.UUID) error {
    query := `
        DELETE FROM organization_members
        WHERE organization_id = $1 AND user_id = $2
    `

    _, err := r.db.Pool.Exec(ctx, query, orgID, userID)
    if err != nil {
        return fmt.Errorf("error deleting member: %w", err)
    }

    return nil
}