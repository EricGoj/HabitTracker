package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

// TestDailyReminderScheduling prueba que el scheduler programa correctamente el job
func TestDailyReminderScheduling(t *testing.T) {
	// Crear scheduler con timezone local
	sched, err := NewScheduler("America/Argentina/Buenos_Aires")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Programar para las 9:00 AM
	timeStr := "09:00"

	callback := func() {
		t.Log("Callback would execute at scheduled time")
	}

	// Programar el recordatorio
	err = sched.ScheduleDailyReminder(timeStr, callback)
	if err != nil {
		t.Fatalf("Failed to schedule reminder: %v", err)
	}

	// Verificar que el scheduler tiene el job programado
	entries := sched.cron.Entries()
	if len(entries) == 0 {
		t.Error("No jobs were scheduled")
	}

	if len(entries) > 0 {
		t.Logf("Job scheduled successfully for next run at: %v", entries[0].Next)
	}
}

// TestImmediateExecution prueba la ejecución inmediata usando un job de prueba
func TestImmediateExecution(t *testing.T) {
	// Crear un cron scheduler de prueba que se ejecuta cada segundo
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}

	c := cron.New(cron.WithLocation(loc), cron.WithSeconds())

	var executed bool
	var mu sync.Mutex

	// Programar para que se ejecute cada segundo
	_, err = c.AddFunc("* * * * * *", func() {
		mu.Lock()
		executed = true
		mu.Unlock()
		t.Log("Test callback executed")
	})

	if err != nil {
		t.Fatalf("Failed to add function: %v", err)
	}

	c.Start()
	defer c.Stop()

	// Esperar hasta 3 segundos para que se ejecute
	time.Sleep(3 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	if !executed {
		t.Error("Callback was not executed within expected time")
	}
}

// TestSchedulerTimezone prueba que el scheduler respeta la zona horaria configurada
func TestSchedulerTimezone(t *testing.T) {
	timezone := "America/Argentina/Buenos_Aires"
	sched, err := NewScheduler(timezone)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	if sched.timezone.String() != timezone {
		t.Errorf("Expected timezone %s, got %s", timezone, sched.timezone.String())
	}
}

// TestInvalidTimeFormat prueba que el scheduler maneja formatos de tiempo inválidos
func TestInvalidTimeFormat(t *testing.T) {
	sched, err := NewScheduler("America/Argentina/Buenos_Aires")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	callback := func() {}

	// Intentar programar con formato inválido
	err = sched.ScheduleDailyReminder("invalid-time", callback)
	if err == nil {
		t.Error("Expected error for invalid time format, got nil")
	}
}

// TestMultipleScheduledJobs prueba que se pueden programar múltiples jobs
func TestMultipleScheduledJobs(t *testing.T) {
	sched, err := NewScheduler("America/Argentina/Buenos_Aires")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	callback1 := func() {
		t.Log("Callback 1 would execute")
	}

	callback2 := func() {
		t.Log("Callback 2 would execute")
	}

	// Programar dos jobs diferentes
	err = sched.ScheduleDailyReminder("09:00", callback1)
	if err != nil {
		t.Fatalf("Failed to schedule first reminder: %v", err)
	}

	err = sched.ScheduleDailyReminder("21:00", callback2)
	if err != nil {
		t.Fatalf("Failed to schedule second reminder: %v", err)
	}

	sched.Start()
	defer sched.Stop()

	// Verificar que ambos jobs están programados
	entries := sched.cron.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 scheduled jobs, got %d", len(entries))
	}

	for i, entry := range entries {
		t.Logf("Job %d scheduled for next run at: %v", i+1, entry.Next)
	}
}

// TestCronExpressionFormat prueba que el formato de expresión cron es correcto
func TestCronExpressionFormat(t *testing.T) {
	testCases := []struct {
		timeStr      string
		expectedHour int
		expectedMin  int
	}{
		{"09:00", 9, 0},
		{"14:30", 14, 30},
		{"00:00", 0, 0},
		{"23:59", 23, 59},
	}

	for _, tc := range testCases {
		t.Run(tc.timeStr, func(t *testing.T) {
			sched, err := NewScheduler("America/Argentina/Buenos_Aires")
			if err != nil {
				t.Fatalf("Failed to create scheduler: %v", err)
			}

			callback := func() {}

			err = sched.ScheduleDailyReminder(tc.timeStr, callback)
			if err != nil {
				t.Fatalf("Failed to schedule reminder for %s: %v", tc.timeStr, err)
			}

			sched.Start()
			defer sched.Stop()

			entries := sched.cron.Entries()
			if len(entries) == 0 {
				t.Error("No jobs were scheduled")
				return
			}

			// Verificar que el próximo tiempo de ejecución tiene la hora y minuto correctos
			next := entries[0].Next
			if next.Hour() != tc.expectedHour || next.Minute() != tc.expectedMin {
				t.Errorf("Expected next run at %02d:%02d, got %02d:%02d",
					tc.expectedHour, tc.expectedMin, next.Hour(), next.Minute())
			}

			t.Logf("Job correctly scheduled for %s, next run: %v", tc.timeStr, next)
		})
	}
}

// TestInvalidTimezone prueba el manejo de zonas horarias inválidas
func TestInvalidTimezone(t *testing.T) {
	_, err := NewScheduler("Invalid/Timezone")
	if err == nil {
		t.Error("Expected error for invalid timezone, got nil")
	}
}
