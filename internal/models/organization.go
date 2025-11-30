package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
		ID          uuid.UUID  `json:"id"`
    ClerkOrgID  string     `json:"clerk_org_id"`
    Name        string     `json:"name"`
    Slug        string     `json:"slug"`
    Description string     `json:"description"`
    LogoURL     string     `json:"logo_url"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type OrganizationRole string

const (
		RoleOwner  OrganizationRole = "owner"
    RoleAdmin  OrganizationRole = "admin"
    RoleMember OrganizationRole = "member"
)

type OrganizationMember struct {
		ID                uuid.UUID        `json:"id"`
    OrganizationID    uuid.UUID        `json:"organization_id"`
    UserID            uuid.UUID        `json:"user_id"`
    Role              OrganizationRole `json:"role"`
    ClerkMembershipID string           `json:"clerk_membership_id"`
    JoinedAt          time.Time        `json:"joined_at"`
    UpdatedAt         time.Time        `json:"updated_at"`
}

type OrganizationWithRole struct {
	Organization
	Role OrganizationRole `json:"role"`
}

type CreateOrganizationRequest struct {
		ClerkOrgID  string `json:"clerk_org_id" validate:"required"`
    Name        string `json:"name" validate:"required"`
    Slug        string `json:"slug" validate:"required"`
    Description string `json:"description"`
    LogoURL     string `json:"logo_url"`
}