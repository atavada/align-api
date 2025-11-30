package routes

import (
	"github.com/atavada/project-management-saas/internal/handlers"
	"github.com/atavada/project-management-saas/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handlers struct {
	Webhook *handlers.WebhookHandler
	User *handlers.UserHandler
	Organization *handlers.OrganizationHandler
}

func SetupRoutes(app *fiber.App, h *Handlers, clerkSecretKey string) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Webhook (no auth required)
	webhooks := api.Group("/webhooks")
	webhooks.Post("/clerk", h.Webhook.HandlerClerkWebhook)

	// Protected routes
	protected := api.Group("", middleware.AuthMiddleware(clerkSecretKey))

	// User routes
	users := protected.Group("/users")
	users.Get("/me", h.User.GetCurrentUser)

	// Organization routes
	organization := protected.Group("/organizations")
	organization.Get("/", h.Organization.ListUserOrganizations)
	organization.Get("/:id", h.Organization.GetOrganization)
}