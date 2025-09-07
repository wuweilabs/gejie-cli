package utils

import (
	"fmt"
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
	fmt.Printf("[TIMING]: %s took %dm %ds %dms\n", t.name, minutes, seconds, milliseconds)
	return elapsed
}

func (t *Timer) LogElapsedWithMessage(message string) time.Duration {
	elapsed := t.Elapsed()
	fmt.Printf("[TIMING] %s: %s took %dms\n", t.name, message, elapsed.Milliseconds())
	return elapsed
}
