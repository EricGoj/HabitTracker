package bot

import (
	"encoding/json"
	"fmt"
	"habittracker/habits"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	habitManager *habits.HabitManager
	userChatID   int64
}

func NewBot(token string, habitManager *habits.HabitManager) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:          api,
		habitManager: habitManager,
	}, nil
}

// Start inicia el bot y comienza a escuchar mensajes
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3600 // 1 hora de timeout para long polling

	updates := b.api.GetUpdatesChan(u)

	log.Println("Bot started, waiting for messages...")

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}
}

// handleMessage maneja los mensajes de texto
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Guardar el chat ID del usuario para notificaciones
	if b.userChatID == 0 {
		b.userChatID = message.Chat.ID
		log.Printf("User chat ID saved: %d", b.userChatID)
	}

	if !message.IsCommand() {
		return
	}

	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.handleHelp(message)
	case "addhabit":
		b.handleAddHabit(message)
	case "listhabits":
		b.handleListHabits(message)
	case "deletehabit":
		b.handleDeleteHabit(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Comando no reconocido. Usa /help para ver los comandos disponibles.")
		b.api.Send(msg)
	}
}

// handleStart maneja el comando /start
func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := "¡Bienvenido al Habit Tracker Bot! 🎯\n\n" +
		"Este bot te ayudará a rastrear tus hábitos diarios.\n" +
		"📅 *Rutina Diaria:*\n" +
		"🌅 08:00 AM - Planificación del día\n" +
		"🌙 09:00 PM - Revisión de progreso\n\n" +
		"Usa /help para ver todos los comandos disponibles."

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

// handleHelp maneja el comando /help
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := "📋 *Comandos disponibles:*\n\n" +
		"/start - Iniciar el bot\n" +
		"/help - Mostrar esta ayuda\n" +
		"/addhabit <nombre> - Agregar un nuevo hábito\n" +
		"/listhabits - Listar todos tus hábitos\n" +
		"/deletehabit <id> - Eliminar un hábito\n\n" +
		"💡 *Ejemplo:*\n" +
		"`/addhabit Hacer ejercicio`"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleAddHabit maneja el comando /addhabit
func (b *Bot) handleAddHabit(message *tgbotapi.Message) {
	args := message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Por favor proporciona un nombre para el hábito.\nEjemplo: /addhabit Hacer ejercicio")
		b.api.Send(msg)
		return
	}

	habit, err := b.habitManager.AddHabit(args, "")
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Error al agregar hábito: %v", err))
		b.api.Send(msg)
		return
	}

	text := fmt.Sprintf("✅ Hábito agregado exitosamente!\n\nID: %d\nNombre: %s", habit.ID, habit.Name)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

// handleListHabits maneja el comando /listhabits
func (b *Bot) handleListHabits(message *tgbotapi.Message) {
	habits := b.habitManager.GetHabits()

	if len(habits) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "No tienes hábitos configurados aún.\nUsa /addhabit para agregar uno.")
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString("📋 *Tus hábitos:*\n\n")

	for _, habit := range habits {
		text.WriteString(fmt.Sprintf("*ID %d:* %s\n", habit.ID, habit.Name))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleDeleteHabit maneja el comando /deletehabit
func (b *Bot) handleDeleteHabit(message *tgbotapi.Message) {
	args := message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Por favor proporciona el ID del hábito a eliminar.\nEjemplo: /deletehabit 1")
		b.api.Send(msg)
		return
	}

	id, err := strconv.Atoi(args)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "ID inválido. Debe ser un número.")
		b.api.Send(msg)
		return
	}

	if err := b.habitManager.DeleteHabit(id); err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Error: %v", err))
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Hábito eliminado exitosamente!")
	b.api.Send(msg)
}

// handleCallback maneja las respuestas de los botones inline
func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	parts := strings.Split(data, "_")

	// Esperamos formato: type_action_id (ej: plan_yes_1, review_no_2)
	// O legacy: action_id (ej: yes_1) -> lo tratamos como review por compatibilidad si es necesario, o lo ignoramos.

	if len(parts) < 2 {
		return
	}

	var actionType, response string
	var habitID int
	var err error

	if len(parts) == 3 {
		actionType = parts[0] // plan o review
		response = parts[1]   // yes o no
		habitID, err = strconv.Atoi(parts[2])
	} else if len(parts) == 2 {
		// Legacy support (asumimos review)
		actionType = "review"
		response = parts[0]
		habitID, err = strconv.Atoi(parts[1])
	}

	if err != nil {
		return
	}

	// Obtener el nombre del hábito
	habits := b.habitManager.GetHabits()
	var habitName string
	for _, h := range habits {
		if h.ID == habitID {
			habitName = h.Name
			break
		}
	}

	var responseText string

	if actionType == "plan" {
		planned := response == "yes"
		if err := b.habitManager.RecordPlan(habitID, planned); err != nil {
			log.Printf("Error recording plan: %v", err)
			return
		}

		if planned {
			responseText = fmt.Sprintf("👍 Planeado: '%s'", habitName)
		} else {
			responseText = fmt.Sprintf("⏭️ Saltado por hoy: '%s'", habitName)
		}

	} else if actionType == "review" {
		completed := response == "yes"
		if err := b.habitManager.RecordCompletion(habitID, completed); err != nil {
			log.Printf("Error recording completion: %v", err)
			return
		}

		if completed {
			responseText = fmt.Sprintf("✅ Completado: '%s'", habitName)
		} else {
			responseText = fmt.Sprintf("❌ No completado: '%s'", habitName)
		}
	}

	// Responder al callback
	callbackConfig := tgbotapi.NewCallback(callback.ID, responseText)
	b.api.Send(callbackConfig)

	// Actualizar el mensaje original para quitar los botones y mostrar la elección
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, responseText)
	b.api.Send(edit)
}

// SendMorningGreeting envía el saludo matutino y pregunta qué hábitos se harán hoy
func (b *Bot) SendMorningGreeting() error {
	if b.userChatID == 0 {
		log.Println("No user chat ID available yet, skipping morning greeting")
		return nil
	}

	habits := b.habitManager.GetHabits()

	if len(habits) == 0 {
		msg := tgbotapi.NewMessage(b.userChatID, "No tienes hábitos configurados. Usa /addhabit para agregar uno.")
		_, err := b.api.Send(msg)
		return err
	}

	text := "🌅 *Buenos días!* Planifiquemos tu día.\n¿Qué hábitos harás hoy?"
	msg := tgbotapi.NewMessage(b.userChatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)

	// Enviar un mensaje por cada hábito con botones de planificación
	for _, habit := range habits {
		habitText := fmt.Sprintf("🎯 *%s*", habit.Name)
		habitMsg := tgbotapi.NewMessage(b.userChatID, habitText)
		habitMsg.ParseMode = "Markdown"

		// Crear botones inline
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👍 Lo haré", fmt.Sprintf("plan_yes_%d", habit.ID)),
				tgbotapi.NewInlineKeyboardButtonData("⏭️ Hoy no", fmt.Sprintf("plan_no_%d", habit.ID)),
			),
		)
		habitMsg.ReplyMarkup = keyboard

		if _, err := b.api.Send(habitMsg); err != nil {
			log.Printf("Error sending habit planner: %v", err)
		}
	}

	return nil
}

// SendEveningReview envía la revisión nocturna de los hábitos planeados
func (b *Bot) SendEveningReview() error {
	if b.userChatID == 0 {
		log.Println("No user chat ID available yet, skipping evening review")
		return nil
	}

	// Obtener planes de hoy
	now := time.Now()
	date := now.Format("2006-01-02")
	dailyPlans := b.habitManager.GetDailyPlans(date)
	allHabits := b.habitManager.GetHabits()

	// Mapa para acceso rápido a hábitos
	habitMap := make(map[int]string)
	for _, h := range allHabits {
		habitMap[h.ID] = h.Name
	}

	// Filtrar hábitos que se planearon hacer (o todos si no hubo planificación explícita, decisión de diseño)
	// Por ahora, solo preguntamos por los que dijeron "SI" o los que no respondieron (asumimos que quizás lo hicieron)
	// O simplificamos: preguntamos por TODOS los hábitos activos, pero personalizamos el mensaje si dijeron que NO.
	// Vamos a preguntar por los que dijeron SI o no respondieron.

	var habitsToReview []int
	plannedMap := make(map[int]bool)

	for _, plan := range dailyPlans {
		plannedMap[plan.HabitID] = plan.Planned
	}

	for _, h := range allHabits {
		planned, responded := plannedMap[h.ID]
		// Si dijo que SI (planned=true) O no respondió (!responded), preguntamos.
		// Si dijo que NO (planned=false), no preguntamos (respetamos su decisión matutina).
		if !responded || planned {
			habitsToReview = append(habitsToReview, h.ID)
		}
	}

	if len(habitsToReview) == 0 {
		msg := tgbotapi.NewMessage(b.userChatID, "🌙 *Buenas noches!* Hoy no planificaste ningún hábito. ¡Mañana será otro día!")
		msg.ParseMode = "Markdown"
		b.api.Send(msg)
		return nil
	}

	text := "🌙 *Buenas noches!* Es hora de revisar tu progreso de hoy."
	msg := tgbotapi.NewMessage(b.userChatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)

	for _, habitID := range habitsToReview {
		name := habitMap[habitID]
		habitText := fmt.Sprintf("❓ *%s*\n¿Lo completaste?", name)
		habitMsg := tgbotapi.NewMessage(b.userChatID, habitText)
		habitMsg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Sí", fmt.Sprintf("review_yes_%d", habitID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ No", fmt.Sprintf("review_no_%d", habitID)),
			),
		)
		habitMsg.ReplyMarkup = keyboard

		if _, err := b.api.Send(habitMsg); err != nil {
			log.Printf("Error sending habit review: %v", err)
		}
	}

	return nil
}

// GetUserChatID devuelve el chat ID del usuario
func (b *Bot) GetUserChatID() int64 {
	return b.userChatID
}

// SetUserChatID configura manualmente el chat ID del usuario (útil para testing)
func (b *Bot) SetUserChatID(chatID int64) {
	b.userChatID = chatID
	log.Printf("User chat ID set manually: %d", chatID)
}

// SetWebhook configura el webhook de Telegram
func (b *Bot) SetWebhook(webhookURL string) error {
	webhookConfig, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}

	_, err = b.api.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	info, err := b.api.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %w", err)
	}

	log.Printf("Webhook set successfully!")
	log.Printf("URL: %s", info.URL)
	log.Printf("Pending updates: %d", info.PendingUpdateCount)

	return nil
}

// GetWebhookHandler retorna un http.Handler para procesar actualizaciones del webhook
func (b *Bot) GetWebhookHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("📨 Webhook request received")
		log.Printf("Method: %s | Path: %s", r.Method, r.URL.Path)
		log.Printf("Remote Address: %s", r.RemoteAddr)

		// Parsear el update de Telegram
		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("❌ Error decoding update: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Log del update completo en JSON
		updateJSON, _ := json.MarshalIndent(update, "", "  ")
		log.Printf("📦 Full Update JSON:\n%s", string(updateJSON))

		// Log de información específica
		log.Printf("Update ID: %d", update.UpdateID)

		if update.Message != nil {
			log.Printf("📩 Message received:")
			log.Printf("  From: %s (@%s) [ID: %d]",
				update.Message.From.FirstName,
				update.Message.From.UserName,
				update.Message.From.ID)
			log.Printf("  Chat ID: %d", update.Message.Chat.ID)
			log.Printf("  Text: %s", update.Message.Text)
			if update.Message.IsCommand() {
				log.Printf("  Command: /%s", update.Message.Command())
			}
		}

		if update.CallbackQuery != nil {
			log.Printf("🔘 Callback Query received:")
			log.Printf("  From: %s (@%s) [ID: %d]",
				update.CallbackQuery.From.FirstName,
				update.CallbackQuery.From.UserName,
				update.CallbackQuery.From.ID)
			log.Printf("  Data: %s", update.CallbackQuery.Data)
		}

		// Procesar el update
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}

		log.Println("✅ Update processed successfully")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		w.WriteHeader(http.StatusOK)
	}
}
