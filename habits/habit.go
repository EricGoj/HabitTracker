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

type HabitManager struct {
	habits          []Habit
	responses       []HabitResponse
	habitsFile      string
	responsesFile   string
	mu              sync.RWMutex
	nextID          int
}

func NewHabitManager(habitsFile, responsesFile string) *HabitManager {
	hm := &HabitManager{
		habits:        []Habit{},
		responses:     []HabitResponse{},
		habitsFile:    habitsFile,
		responsesFile: responsesFile,
		nextID:        1,
	}
	hm.LoadHabits()
	hm.LoadResponses()
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
