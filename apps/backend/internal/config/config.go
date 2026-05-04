package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	// Google
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string
	GOOGLE_REDIRECT_URL  string

	// JWT
	JwtExpires int
	JwtSecret  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv := getEnv("APP_ENV", "development")
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		if appEnv != "development" {
			return nil, fmt.Errorf("DATABASE_URL is required in %s environment", appEnv)
		}
		dbURL = "postgres://postgres:password@127.0.0.1:5433/omni-bot?sslmode=disable"
	}

	jwtHours, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN_HOURS", "24"))
	if err != nil {
		log.Println("invalid JWT_EXPIRES_IN_HOURS, defaulting to 24")
		jwtHours = 24
	}

	jwtSecret := getEnv("JWT_SECRET", "")

	if jwtSecret == "" {
		if appEnv == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		log.Println("WARNING: using insecure default JWT_SECRET for development")
		jwtSecret = "dev-secret-do-not-use-in-production"
	}

	return &Config{
		AppEnv:                appEnv,
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           dbURL,
		TelegramBotToken:      mustEnv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: mustEnv("TELEGRAM_WEBHOOK_SECRET"),
		GeminiAPIKey:          mustEnv("GEMINI_API_KEY"),
		GOOGLE_CLIENT_ID:      mustEnv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET:  mustEnv("GOOGLE_CLIENT_SECRET"),
		GOOGLE_REDIRECT_URL:   mustEnv("GOOGLE_REDIRECT_URL"),
		JwtExpires:            jwtHours,
		JwtSecret:             jwtSecret,
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
