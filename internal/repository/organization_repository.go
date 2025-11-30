package repository

import (
	"context"
	"fmt"

	"github.com/atavada/project-management-saas/internal/database"
	"github.com/atavada/project-management-saas/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrganizationRepository struct {
	db *database.DB
}

func NewOrganizationRepository(db *database.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *models.CreateOrganizationRequest) (*models.Organization, error) {
	query := `
		INSERT INTO organizations (clerk_org_id, name, slug, description, logo_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, clerk_org_id, name, slug, description, logo_url, created_at, updated_at
	`

	var result models.Organization
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		org.ClerkOrgID,
		org.Name,
		org.Slug,
		org.Description,
		org.LogoURL,
	).Scan(
		&result.ID,
		&result.ClerkOrgID,
		&result.Name,
		&result.Slug,
		&result.Description,
		&result.LogoURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating organization: %w", err)
	}

	return &result, nil
}

func (r *OrganizationRepository) GetByClerkID(ctx context.Context, clerkOrgID string) (*models.Organization, error) {
    query := `
        SELECT id, clerk_org_id, name, slug, description, logo_url, created_at, updated_at
        FROM organizations
        WHERE clerk_org_id = $1
    `

    var org models.Organization
    err := r.db.Pool.QueryRow(ctx, query, clerkOrgID).Scan(
        &org.ID,
        &org.ClerkOrgID,
        &org.Name,
        &org.Slug,
        &org.Description,
        &org.LogoURL,
        &org.CreatedAt,
        &org.UpdatedAt,
    )

    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("error getting organization: %w", err)
    }

    return &org, nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
    query := `
        SELECT id, clerk_org_id, name, slug, description, logo_url, created_at, updated_at
        FROM organizations
        WHERE id = $1
    `

    var org models.Organization
    err := r.db.Pool.QueryRow(ctx, query, id).Scan(
        &org.ID,
        &org.ClerkOrgID,
        &org.Name,
        &org.Slug,
        &org.Description,
        &org.LogoURL,
        &org.CreatedAt,
        &org.UpdatedAt,
    )

    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("error getting organization: %w", err)
    }

    return &org, nil
}

func (r *OrganizationRepository) Upsert(ctx context.Context, org *models.CreateOrganizationRequest) (*models.Organization, error) {
    query := `
        INSERT INTO organizations (clerk_org_id, name, slug, description, logo_url)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (clerk_org_id) 
        DO UPDATE SET
            name = EXCLUDED.name,
            slug = EXCLUDED.slug,
            description = EXCLUDED.description,
            logo_url = EXCLUDED.logo_url,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, clerk_org_id, name, slug, description, logo_url, created_at, updated_at
    `

    var result models.Organization
    err := r.db.Pool.QueryRow(
        ctx,
        query,
        org.ClerkOrgID,
        org.Name,
        org.Slug,
        org.Description,
        org.LogoURL,
    ).Scan(
        &result.ID,
        &result.ClerkOrgID,
        &result.Name,
        &result.Slug,
        &result.Description,
        &result.LogoURL,
        &result.CreatedAt,
        &result.UpdatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("error upserting organization: %w", err)
    }

    return &result, nil
}

func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.OrganizationWithRole, error) {
    query := `
        SELECT 
            o.id, o.clerk_org_id, o.name, o.slug, o.description, o.logo_url, 
            o.created_at, o.updated_at, om.role
        FROM organizations o
        INNER JOIN organization_members om ON o.id = om.organization_id
        WHERE om.user_id = $1
        ORDER BY om.joined_at DESC
    `

    rows, err := r.db.Pool.Query(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error getting user organizations: %w", err)
    }
    defer rows.Close()

    var organizations []models.OrganizationWithRole
    for rows.Next() {
        var org models.OrganizationWithRole
        err := rows.Scan(
            &org.ID,
            &org.ClerkOrgID,
            &org.Name,
            &org.Slug,
            &org.Description,
            &org.LogoURL,
            &org.CreatedAt,
            &org.UpdatedAt,
            &org.Role,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning organization: %w", err)
        }
        organizations = append(organizations, org)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating organizations: %w", err)
    }

    return organizations, nil
}