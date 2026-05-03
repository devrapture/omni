package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string

	// Telegram
	TelegramBotToken      string
	TelegramWebhookSecret string

	// Gemini
	GeminiAPIKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv := getEnv("APP_ENV", "development")
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		if appEnv != "development" {
			return nil, fmt.Errorf("DATABASE_URL is required in %s environment", appEnv)
		}
		dbURL = "postgres://postgres:password@localhost:5432/omni-bot?sslmode=disable"
	}

	return &Config{
		AppEnv:                appEnv,
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           dbURL,
		TelegramBotToken:      mustEnv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: mustEnv("TELEGRAM_WEBHOOK_SECRET"),
		GeminiAPIKey:          mustEnv("GEMINI_API_KEY"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return val
}
