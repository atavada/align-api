package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL 				string
	Port								string
	ClerkSecretKey 			string
	ClerkPublishableKey string
	ClerkWebhookSecret  string
	InngestEventKey     string
	InngestBaseURL      string
	AllowedOrigins      string
	Environment         string
}

func Load() (*Config, error) {
	// Load .env file
	if os.Getenv("ENV") != "production" {
		godotenv.Load()
	}

	config := &Config{
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		Port:                getEnv("PORT", "8080"),
		ClerkSecretKey:      getEnv("CLERK_SECRET_KEY", ""),
		ClerkPublishableKey: getEnv("CLERK_PUBLISHABLE_KEY", ""),
		ClerkWebhookSecret:  getEnv("CLERK_WEBHOOK_SECRET", ""),
		InngestEventKey:     getEnv("INNGEST_EVENT_KEY", ""),
		InngestBaseURL:      getEnv("INNGEST_BASE_URL", "https://inn.gs"),
		AllowedOrigins:      getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		Environment:         getEnv("ENV", "development"),
	}

	return config, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}