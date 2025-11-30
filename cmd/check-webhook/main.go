package main

import (
	"encoding/json"
	"fmt"
	"habittracker/config"
	"log"
	"net/http"
	"os"
)

// Script para verificar el estado del webhook en Telegram
func main() {
	// Cargar configuración
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	token := config.AppConfig.TelegramBotToken

	// Consultar información del webhook
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", token)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error getting webhook info: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			URL                  string `json:"url"`
			HasCustomCertificate bool   `json:"has_custom_certificate"`
			PendingUpdateCount   int    `json:"pending_update_count"`
			LastErrorDate        int    `json:"last_error_date,omitempty"`
			LastErrorMessage     string `json:"last_error_message,omitempty"`
			MaxConnections       int    `json:"max_connections,omitempty"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	if !result.Ok {
		fmt.Println("❌ Error getting webhook info")
		os.Exit(1)
	}

	fmt.Println("📡 Webhook Status:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if result.Result.URL == "" {
		fmt.Println("❌ No webhook configured (using long polling)")
	} else {
		fmt.Printf("✅ Webhook URL: %s\n", result.Result.URL)
		fmt.Printf("⏳ Pending updates: %d\n", result.Result.PendingUpdateCount)

		if result.Result.LastErrorDate > 0 {
			fmt.Printf("⚠️  Last error: %s\n", result.Result.LastErrorMessage)
		} else {
			fmt.Println("✅ No errors")
		}
	}
}
