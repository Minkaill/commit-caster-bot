package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramChannelID string
	GroqAPIKey       string
	GitHubSecret     string
	Port             string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChannelID: getEnv("TELEGRAM_CHANNEL_ID", ""),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		GitHubSecret:      getEnv("GITHUB_WEBHOOK_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
