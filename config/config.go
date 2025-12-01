package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	NotificationTime string // Deprecated: Use MorningTime and EveningTime
	MorningTime      string
	EveningTime      string
	Timezone         string
	WebhookURL       string
	Port             string
}

var AppConfig *Config

// LoadConfig carga la configuración desde el archivo .env
func LoadConfig() error {
	// Cargar archivo .env
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	AppConfig = &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		NotificationTime: os.Getenv("NOTIFICATION_TIME"),
		MorningTime:      os.Getenv("MORNING_TIME"),
		EveningTime:      os.Getenv("EVENING_TIME"),
		Timezone:         os.Getenv("TIMEZONE"),
		WebhookURL:       os.Getenv("WEBHOOK_URL"),
		Port:             os.Getenv("PORT"),
	}

	// Validar configuración requerida
	if AppConfig.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if AppConfig.NotificationTime == "" {
		AppConfig.NotificationTime = "09:00"
	}

	if AppConfig.Timezone == "" {
		AppConfig.Timezone = "America/Argentina/Buenos_Aires"
	}

	if AppConfig.MorningTime == "" {
		AppConfig.MorningTime = "08:00"
	}

	if AppConfig.EveningTime == "" {
		AppConfig.EveningTime = "21:00"
	}

	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	return nil
}
