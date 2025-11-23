package main

import (
	"habittracker/bot"
	"habittracker/config"
	"habittracker/habits"
	"habittracker/scheduler"
	"log"
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
	habitManager := habits.NewHabitManager("data/habits.json", "data/responses.json")
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

	// Programar recordatorio diario
	if err := sched.ScheduleDailyReminder(config.AppConfig.NotificationTime, func() {
		if err := telegramBot.SendDailyReminder(); err != nil {
			log.Printf("Error sending daily reminder: %v", err)
		}
	}); err != nil {
		log.Fatalf("Error scheduling daily reminder: %v", err)
	}

	// Iniciar el scheduler
	sched.Start()

	// Iniciar el bot en una goroutine
	go telegramBot.Start()

	log.Println("✅ Habit Tracker Bot is running!")
	log.Printf("📅 Daily reminders scheduled at %s (%s)", config.AppConfig.NotificationTime, config.AppConfig.Timezone)
	log.Println("Press Ctrl+C to stop")

	// Esperar señal de interrupción
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\n🛑 Shutting down gracefully...")
	sched.Stop()
	log.Println("Goodbye!")
}
