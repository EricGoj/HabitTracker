package bot

import (
	"fmt"
	"habittracker/habits"
	"log"
	"strconv"
	"strings"

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
	u.Timeout = 60

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
		"Recibirás notificaciones todos los días a las 9:00 AM para revisar tus hábitos.\n\n" +
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
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 2 {
		return
	}

	action := parts[0]
	habitID, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	completed := action == "yes"

	if err := b.habitManager.RecordResponse(habitID, completed); err != nil {
		log.Printf("Error recording response: %v", err)
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

	emoji := "✅"
	status := "completado"
	if !completed {
		emoji = "❌"
		status = "no completado"
	}

	responseText := fmt.Sprintf("%s Hábito '%s' marcado como %s", emoji, habitName, status)

	// Responder al callback
	callbackConfig := tgbotapi.NewCallback(callback.ID, responseText)
	b.api.Send(callbackConfig)

	// Actualizar el mensaje
	editText := fmt.Sprintf("Respuesta registrada: %s", responseText)
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, editText)
	b.api.Send(edit)
}

// SendDailyReminder envía el recordatorio diario con los hábitos
func (b *Bot) SendDailyReminder() error {
	if b.userChatID == 0 {
		log.Println("No user chat ID available yet, skipping reminder")
		return nil
	}

	habits := b.habitManager.GetHabits()

	if len(habits) == 0 {
		msg := tgbotapi.NewMessage(b.userChatID, "No tienes hábitos configurados. Usa /addhabit para agregar uno.")
		_, err := b.api.Send(msg)
		return err
	}

	text := "🌅 *Buenos días!* Es hora de revisar tus hábitos de hoy:\n\n"
	msg := tgbotapi.NewMessage(b.userChatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)

	// Enviar un mensaje por cada hábito con botones
	for _, habit := range habits {
		habitText := fmt.Sprintf("*%s*\n¿Completaste este hábito?", habit.Name)
		habitMsg := tgbotapi.NewMessage(b.userChatID, habitText)
		habitMsg.ParseMode = "Markdown"

		// Crear botones inline
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Sí", fmt.Sprintf("yes_%d", habit.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ No", fmt.Sprintf("no_%d", habit.ID)),
			),
		)
		habitMsg.ReplyMarkup = keyboard

		if _, err := b.api.Send(habitMsg); err != nil {
			log.Printf("Error sending habit reminder: %v", err)
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
