package middleware

import (
	"context"
	"strings"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(clerkSecretKey string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := tokenParts[1]

		// Verify token with Clerk
		clerk.SetKey(clerkSecretKey)
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store user info in context
		c.Locals("clerkUserID", claims.Subject)
		if claims.ActiveOrganizationID != "" {
			c.Locals("clerkOrgID", claims.ActiveOrganizationID)
		}

		return c.Next()
	}
}