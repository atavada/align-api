package main

import (
	"log"

	"github.com/atavada/project-management-saas/internal/config"
	"github.com/atavada/project-management-saas/internal/database"
	"github.com/atavada/project-management-saas/internal/handlers"
	"github.com/atavada/project-management-saas/internal/repository"
	"github.com/atavada/project-management-saas/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to DB
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repo
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	memberRepo := repository.NewOrganizationMemberRepository(db)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(
		userRepo,
		orgRepo,
		memberRepo,
		cfg.ClerkWebhookSecret,
	)
	userHandler := handlers.NewUserHandler(userRepo)
	orgHandler := handlers.NewOrganizationHandler(userRepo, orgRepo, memberRepo)

	allHandlers := &routes.Handlers{
		Webhook: webhookHandler,
		User: userHandler,
		Organization: orgHandler,
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.AllowedOrigins},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(app, allHandlers, cfg.ClerkSecretKey)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"message": err.Error(),
	})
}