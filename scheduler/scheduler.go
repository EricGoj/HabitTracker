package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron     *cron.Cron
	timezone *time.Location
}

// NewScheduler crea un nuevo scheduler
func NewScheduler(timezone string) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	c := cron.New(cron.WithLocation(loc))

	return &Scheduler{
		cron:     c,
		timezone: loc,
	}, nil
}

// ScheduleDailyReminder programa un recordatorio diario
func (s *Scheduler) ScheduleDailyReminder(timeStr string, callback func()) error {
	// Parsear la hora (formato HH:MM)
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return err
	}

	// Crear expresión cron (minuto hora * * *)
	cronExpr := fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour())

	log.Printf("Scheduling daily reminder at %s (cron: %s) in timezone %s", timeStr, cronExpr, s.timezone)

	_, err = s.cron.AddFunc(cronExpr, func() {
		log.Println("Executing daily reminder...")
		callback()
	})

	return err
}

// Start inicia el scheduler
func (s *Scheduler) Start() {
	log.Println("Scheduler started")
	s.cron.Start()
}

// Stop detiene el scheduler
func (s *Scheduler) Stop() {
	log.Println("Scheduler stopped")
	s.cron.Stop()
}
