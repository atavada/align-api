package handlers

import (
	"context"

	"github.com/atavada/project-management-saas/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GetCurrentUser returns the authenticated user's profile
func (h *UserHandler) GetCurrentUser(c fiber.Ctx) error {
	ctx := context.Background()
	clerkUserID := c.Locals("clerkUserID").(string)

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

	return c.JSON(fiber.Map{
		"data": user,
	})
}