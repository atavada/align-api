package handlers

import (
	"context"

	"github.com/atavada/project-management-saas/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	userRepo *repository.UserRepository
	orgRepo *repository.OrganizationRepository
	memberRepo *repository.OrganizationMemberRepository
}

func NewOrganizationHandler(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	memberRepo *repository.OrganizationMemberRepository,
) *OrganizationHandler {
	return &OrganizationHandler{
		userRepo: userRepo,
		orgRepo: orgRepo,
		memberRepo: memberRepo,
	}
}

// ListUserOrganizations returns all organizations the uis is a member of
func (h *OrganizationHandler) ListUserOrganizations(c fiber.Ctx) error {
	ctx := context.Background()
	clerkUserID := c.Locals("clerkUserID").(string)

	// Get user from DB
	user, err := h.userRepo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get user's organization
	organization, err := h.orgRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch organization",
		})
	}

	return c.JSON(fiber.Map{
		"data": organization,
	})
}

// GetOrganization returns details of a specific organization
func (h *OrganizationHandler) GetOrganization(c fiber.Ctx) error {
	ctx := context.Background()

	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	clerkUserID := c.Locals("clerkUserID").(string)

	// Get user
	user, err := h.userRepo.GetByClerkID(ctx, clerkUserID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if user is member of the organization
	member, err := h.memberRepo.GetMember(ctx, orgID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify membership",
		})
	}
	if member == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get organization
	org, err := h.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch organization",
		})
	}
	if org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"organization": org,
			"role": member.Role,
		},
	})
}

