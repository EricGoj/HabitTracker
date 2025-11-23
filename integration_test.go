package main

import (
	"habittracker/bot"
	"habittracker/config"
	"habittracker/habits"
	"habittracker/scheduler"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestEndToEndNotification prueba el flujo completo de notificación
// Este test requiere que el archivo .env esté configurado correctamente
// y enviará una notificación real a tu cuenta de Telegram
func TestEndToEndNotification(t *testing.T) {
	// Cargar configuración
	if err := config.LoadConfig(); err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Verificar que tenemos el token configurado
	if config.AppConfig.TelegramBotToken == "" {
		t.Skip("Skipping integration test: TELEGRAM_BOT_TOKEN not configured")
	}

	t.Log("🚀 Iniciando test de integración end-to-end...")
	t.Log("⚠️  Este test enviará una notificación REAL a tu cuenta de Telegram")

	// Crear directorios de datos temporales para el test
	testDataDir := "data/test"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)

	// Inicializar el gestor de hábitos con archivos de prueba
	habitManager := habits.NewHabitManager(
		testDataDir+"/habits.json",
		testDataDir+"/responses.json",
	)

	// Agregar algunos hábitos de prueba
	habitManager.AddHabit("Test: Hacer ejercicio", "Hábito de prueba")
	habitManager.AddHabit("Test: Meditar", "Hábito de prueba")
	t.Log("✅ Hábitos de prueba creados")

	// Inicializar el bot
	telegramBot, err := bot.NewBot(config.AppConfig.TelegramBotToken, habitManager)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}
	t.Log("✅ Bot inicializado")

	// Configurar el chat ID desde la variable de entorno o usar uno de prueba
	chatID := config.AppConfig.TelegramChatID
	if chatID == "" {
		t.Skip("Skipping test: TELEGRAM_CHAT_ID not configured in .env")
	}

	// Convertir chat ID a int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		t.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	// Configurar el chat ID manualmente para evitar necesitar /start
	telegramBot.SetUserChatID(chatIDInt)
	t.Logf("✅ Chat ID configurado: %d", chatIDInt)

	// NO iniciamos bot.Start() para evitar conflictos con otras instancias
	// El bot solo se usará para enviar mensajes, no para recibirlos
	t.Log("✅ Bot listo para enviar notificaciones")

	// Inicializar el scheduler
	sched, err := scheduler.NewScheduler(config.AppConfig.Timezone)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Programar para que se ejecute en 5 segundos desde ahora
	now := time.Now()
	futureTime := now.Add(5 * time.Second)
	timeStr := futureTime.Format("15:04")

	t.Logf("⏰ Programando notificación para: %s (en 5 segundos)", timeStr)

	// Variable para verificar que el callback se ejecutó
	callbackExecuted := false

	// Programar el recordatorio
	err = sched.ScheduleDailyReminder(timeStr, func() {
		t.Log("📨 Ejecutando callback de notificación...")
		if err := telegramBot.SendDailyReminder(); err != nil {
			t.Logf("Error sending reminder: %v", err)
		} else {
			callbackExecuted = true
			t.Log("✅ Notificación enviada exitosamente!")
		}
	})

	if err != nil {
		t.Fatalf("Failed to schedule reminder: %v", err)
	}

	// Iniciar el scheduler
	sched.Start()
	defer sched.Stop()

	t.Log("⏳ Esperando a que se ejecute el job (esto tomará ~7 segundos)...")
	t.Log("💡 Revisa tu Telegram, deberías recibir una notificación pronto")

	// Esperar más tiempo del programado para asegurar que se ejecute
	time.Sleep(8 * time.Second)

	// Verificar que el callback se ejecutó
	if !callbackExecuted {
		t.Error("❌ El callback no se ejecutó dentro del tiempo esperado")
	} else {
		t.Log("✅ Test completado exitosamente!")
		t.Log("📱 Deberías haber recibido una notificación en Telegram con los hábitos de prueba")
	}
}

// TestEndToEndWithManualTrigger prueba el envío manual de notificación
// sin esperar al scheduler (más rápido)
func TestEndToEndWithManualTrigger(t *testing.T) {
	// Cargar configuración
	if err := config.LoadConfig(); err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	if config.AppConfig.TelegramBotToken == "" {
		t.Skip("Skipping integration test: TELEGRAM_BOT_TOKEN not configured")
	}

	t.Log("🚀 Iniciando test de envío manual de notificación...")
	t.Log("⚠️  Este test enviará una notificación REAL a tu cuenta de Telegram")

	// Crear directorios de datos temporales
	testDataDir := "data/test_manual"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)

	// Inicializar el gestor de hábitos
	habitManager := habits.NewHabitManager(
		testDataDir+"/habits.json",
		testDataDir+"/responses.json",
	)

	// Agregar hábitos de prueba
	habitManager.AddHabit("Test Manual: Leer", "Hábito de prueba")
	habitManager.AddHabit("Test Manual: Escribir", "Hábito de prueba")
	t.Log("✅ Hábitos de prueba creados")

	// Inicializar el bot
	telegramBot, err := bot.NewBot(config.AppConfig.TelegramBotToken, habitManager)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}
	t.Log("✅ Bot inicializado")

	// Configurar el chat ID desde la variable de entorno
	chatID := config.AppConfig.TelegramChatID
	if chatID == "" {
		t.Skip("Skipping test: TELEGRAM_CHAT_ID not configured in .env")
	}

	// Convertir chat ID a int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		t.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	// Configurar el chat ID manualmente
	telegramBot.SetUserChatID(chatIDInt)
	t.Logf("✅ Chat ID configurado: %d", chatIDInt)

	// NO iniciamos bot.Start() para evitar conflictos
	t.Log("✅ Bot listo para enviar notificaciones")

	// Dar un momento para que el bot esté listo
	time.Sleep(1 * time.Second)

	// Enviar la notificación inmediatamente
	t.Log("📨 Enviando notificación manual...")
	err = telegramBot.SendDailyReminder()

	if err != nil {
		t.Fatalf("Error al enviar notificación: %v", err)
	}

	t.Log("✅ Notificación enviada exitosamente!")
	t.Log("📱 Revisa tu Telegram, deberías haber recibido la notificación")

	// Dar tiempo para que el mensaje se envíe
	time.Sleep(1 * time.Second)
}
