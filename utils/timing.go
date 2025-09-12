package utils

import (
	"log/slog"
	"time"
)

type Timer struct {
	start time.Time
	name  string
}

func NewTimer(name string) *Timer {
	return &Timer{
		start: time.Now(),
		name:  name,
	}
}

func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

func (t *Timer) ElapsedMs() int64 {
	return time.Since(t.start).Milliseconds()
}

func (t *Timer) LogElapsed() time.Duration {
	elapsed := t.Elapsed()
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := elapsed.Milliseconds() % 1000
	slog.Info("timing elapsed",
		"name", t.name,
		"minutes", minutes,
		"seconds", seconds,
		"milliseconds", milliseconds,
	)
	return elapsed
}

func (t *Timer) LogElapsedWithMessage(message string) time.Duration {
	elapsed := t.Elapsed()
	slog.Info("timing elapsed",
		"name", t.name,
		"message", message,
		"ms", elapsed.Milliseconds(),
	)
	return elapsed
}
