package habits

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Habit struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type HabitResponse struct {
	HabitID   int       `json:"habit_id"`
	Completed bool      `json:"completed"`
	Date      string    `json:"date"`
	Timestamp time.Time `json:"timestamp"`
}

type DailyLog struct {
	Date      string `json:"date"`
	HabitID   int    `json:"habit_id"`
	Planned   bool   `json:"planned"`
	Completed bool   `json:"completed"`
}

type HabitManager struct {
	habits        []Habit
	responses     []HabitResponse
	dailyLogs     []DailyLog
	habitsFile    string
	responsesFile string
	dailyLogsFile string
	mu            sync.RWMutex
	nextID        int
}

func NewHabitManager(habitsFile, responsesFile, dailyLogsFile string) *HabitManager {
	hm := &HabitManager{
		habits:        []Habit{},
		responses:     []HabitResponse{},
		dailyLogs:     []DailyLog{},
		habitsFile:    habitsFile,
		responsesFile: responsesFile,
		dailyLogsFile: dailyLogsFile,
		nextID:        1,
	}
	hm.LoadHabits()
	hm.LoadResponses()
	hm.LoadDailyLogs()
	return hm
}

// AddHabit agrega un nuevo hábito
func (hm *HabitManager) AddHabit(name, description string) (*Habit, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	habit := Habit{
		ID:          hm.nextID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	hm.habits = append(hm.habits, habit)
	hm.nextID++

	if err := hm.saveHabits(); err != nil {
		return nil, err
	}

	return &habit, nil
}

// GetHabits devuelve todos los hábitos
func (hm *HabitManager) GetHabits() []Habit {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.habits
}

// DeleteHabit elimina un hábito por ID
func (hm *HabitManager) DeleteHabit(id int) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for i, habit := range hm.habits {
		if habit.ID == id {
			hm.habits = append(hm.habits[:i], hm.habits[i+1:]...)
			return hm.saveHabits()
		}
	}

	return fmt.Errorf("habit with ID %d not found", id)
}

// RecordResponse registra una respuesta para un hábito
func (hm *HabitManager) RecordResponse(habitID int, completed bool) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	now := time.Now()
	date := now.Format("2006-01-02")

	response := HabitResponse{
		HabitID:   habitID,
		Completed: completed,
		Date:      date,
		Timestamp: now,
	}

	hm.responses = append(hm.responses, response)
	return hm.saveResponses()
}

// RecordPlan registra si un hábito fue planeado para hoy
func (hm *HabitManager) RecordPlan(habitID int, planned bool) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	now := time.Now()
	date := now.Format("2006-01-02")

	// Buscar si ya existe un log para hoy
	for i, log := range hm.dailyLogs {
		if log.Date == date && log.HabitID == habitID {
			hm.dailyLogs[i].Planned = planned
			return hm.saveDailyLogs()
		}
	}

	// Si no existe, crear uno nuevo
	newLog := DailyLog{
		Date:      date,
		HabitID:   habitID,
		Planned:   planned,
		Completed: false,
	}

	hm.dailyLogs = append(hm.dailyLogs, newLog)
	return hm.saveDailyLogs()
}

// RecordCompletion registra si un hábito fue completado hoy
func (hm *HabitManager) RecordCompletion(habitID int, completed bool) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	now := time.Now()
	date := now.Format("2006-01-02")

	// Buscar si ya existe un log para hoy
	for i, log := range hm.dailyLogs {
		if log.Date == date && log.HabitID == habitID {
			hm.dailyLogs[i].Completed = completed
			return hm.saveDailyLogs()
		}
	}

	// Si no existe, crear uno nuevo (asumiendo que no fue planeado explícitamente pero se hizo)
	newLog := DailyLog{
		Date:      date,
		HabitID:   habitID,
		Planned:   false, // O true? Por ahora false si no hubo plan previo
		Completed: completed,
	}

	hm.dailyLogs = append(hm.dailyLogs, newLog)
	return hm.saveDailyLogs()
}

// GetDailyPlans devuelve los logs del día especificado
func (hm *HabitManager) GetDailyPlans(date string) []DailyLog {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var logs []DailyLog
	for _, log := range hm.dailyLogs {
		if log.Date == date {
			logs = append(logs, log)
		}
	}
	return logs
}

// LoadDailyLogs carga los logs diarios desde el archivo
func (hm *HabitManager) LoadDailyLogs() error {
	data, err := os.ReadFile(hm.dailyLogsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &hm.dailyLogs)
}

// saveDailyLogs guarda los logs diarios en el archivo
func (hm *HabitManager) saveDailyLogs() error {
	data, err := json.MarshalIndent(hm.dailyLogs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(hm.dailyLogsFile, data, 0644)
}

// LoadHabits carga los hábitos desde el archivo
func (hm *HabitManager) LoadHabits() error {
	data, err := os.ReadFile(hm.habitsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Archivo no existe aún, está bien
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &hm.habits); err != nil {
		return err
	}

	// Actualizar nextID
	for _, habit := range hm.habits {
		if habit.ID >= hm.nextID {
			hm.nextID = habit.ID + 1
		}
	}

	return nil
}

// LoadResponses carga las respuestas desde el archivo
func (hm *HabitManager) LoadResponses() error {
	data, err := os.ReadFile(hm.responsesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &hm.responses)
}

// saveHabits guarda los hábitos en el archivo
func (hm *HabitManager) saveHabits() error {
	data, err := json.MarshalIndent(hm.habits, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(hm.habitsFile, data, 0644)
}

// saveResponses guarda las respuestas en el archivo
func (hm *HabitManager) saveResponses() error {
	data, err := json.MarshalIndent(hm.responses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(hm.responsesFile, data, 0644)
}
