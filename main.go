package main

import (
	"fmt"
	"habittracker/bot"
	"habittracker/config"
	"habittracker/habits"
	"habittracker/scheduler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Cargar configuración
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Println("Configuration loaded successfully")

	// Crear directorios de datos si no existen
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Error creating data directory: %v", err)
	}

	// Inicializar el gestor de hábitos
	habitManager := habits.NewHabitManager("data/habits.json", "data/responses.json", "data/daily_logs.json")
	log.Println("Habit manager initialized")

	// Inicializar el bot
	telegramBot, err := bot.NewBot(config.AppConfig.TelegramBotToken, habitManager)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// Inicializar el scheduler
	sched, err := scheduler.NewScheduler(config.AppConfig.Timezone)
	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	// Programar saludo matutino (Planificación)
	if err := sched.ScheduleDailyReminder(config.AppConfig.MorningTime, func() {
		if err := telegramBot.SendMorningGreeting(); err != nil {
			log.Printf("Error sending morning greeting: %v", err)
		}
	}); err != nil {
		log.Fatalf("Error scheduling morning greeting: %v", err)
	}

	// Programar revisión nocturna (Verificación)
	if err := sched.ScheduleDailyReminder(config.AppConfig.EveningTime, func() {
		if err := telegramBot.SendEveningReview(); err != nil {
			log.Printf("Error sending evening review: %v", err)
		}
	}); err != nil {
		log.Fatalf("Error scheduling evening review: %v", err)
	}

	// Iniciar el scheduler
	sched.Start()

	// Configurar modo de operación: webhook o polling
	if config.AppConfig.WebhookURL != "" {
		// Modo webhook
		log.Println("🌐 Starting in WEBHOOK mode")

		// Configurar el webhook en Telegram
		if err := telegramBot.SetWebhook(config.AppConfig.WebhookURL); err != nil {
			log.Fatalf("Error setting webhook: %v", err)
		}

		// Configurar el servidor HTTP
		http.HandleFunc("/", telegramBot.GetWebhookHandler())

		addr := fmt.Sprintf(":%s", config.AppConfig.Port)
		log.Printf("✅ Habit Tracker Bot is running in WEBHOOK mode!")
		log.Printf("📡 Listening on %s", addr)
		log.Printf("🔗 Webhook URL: %s", config.AppConfig.WebhookURL)
		log.Printf("📅 Morning greeting scheduled at %s (%s)", config.AppConfig.MorningTime, config.AppConfig.Timezone)
		log.Printf("📅 Evening review scheduled at %s (%s)", config.AppConfig.EveningTime, config.AppConfig.Timezone)

		// Iniciar servidor HTTP en una goroutine
		go func() {
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatalf("Error starting HTTP server: %v", err)
			}
		}()
	} else {
		// Modo long polling (fallback)
		log.Println("📞 Starting in POLLING mode")

		// Iniciar el bot en una goroutine
		go telegramBot.Start()

		log.Println("✅ Habit Tracker Bot is running in POLLING mode!")
		log.Printf("📅 Morning greeting scheduled at %s (%s)", config.AppConfig.MorningTime, config.AppConfig.Timezone)
		log.Printf("📅 Evening review scheduled at %s (%s)", config.AppConfig.EveningTime, config.AppConfig.Timezone)
	}

	log.Println("Press Ctrl+C to stop")

	// Esperar señal de interrupción
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\n🛑 Shutting down gracefully...")
	sched.Stop()
	log.Println("Goodbye!")
}
