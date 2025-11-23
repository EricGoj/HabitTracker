package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	NotificationTime string
	Timezone         string
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
		Timezone:         os.Getenv("TIMEZONE"),
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

	return nil
}
