// Package timer implements room-owned game timer state.
package timer

import "time"

// DefaultSteps stores Arcturus-compatible timer durations.
var DefaultSteps = [...]time.Duration{30 * time.Second, 60 * time.Second, 120 * time.Second, 180 * time.Second, 300 * time.Second, 600 * time.Second}

// Timer stores one game timer state without owning a goroutine.
type Timer struct {
	// Steps stores selectable durations.
	Steps []time.Duration
	// Step stores the selected duration index.
	Step int
	// Remaining stores time left.
	Remaining time.Duration
	// Running reports whether ticks consume time.
	Running bool
	// Started reports whether a match exists.
	Started bool
}

// New creates a timer with normalized positive steps.
func New(steps []time.Duration) *Timer {
	normalized := make([]time.Duration, 0, len(steps))
	for _, step := range steps {
		if step > 0 {
			normalized = append(normalized, step)
		}
	}
	if len(normalized) == 0 {
		normalized = append(normalized, DefaultSteps[:]...)
	}
	return &Timer{Steps: normalized, Remaining: normalized[0]}
}

// Start begins or resumes the selected duration.
func (timer *Timer) Start() bool {
	if timer.Running {
		return false
	}
	if !timer.Started || timer.Remaining <= 0 {
		timer.Remaining = timer.Steps[timer.Step]
	}
	timer.Started, timer.Running = true, true
	return true
}

// Toggle pauses or resumes an existing match.
func (timer *Timer) Toggle() bool {
	if !timer.Started {
		return timer.Start()
	}
	timer.Running = !timer.Running
	return timer.Running
}

// Increase selects the next configured duration and extends the current match by its delta.
func (timer *Timer) Increase() time.Duration {
	previous := timer.Steps[timer.Step]
	timer.Step = (timer.Step + 1) % len(timer.Steps)
	next := timer.Steps[timer.Step]
	if timer.Started {
		timer.Remaining += next - previous
	} else {
		timer.Remaining = next
	}
	if timer.Remaining < 0 {
		timer.Remaining = 0
	}
	return timer.Remaining
}

// Tick advances state and reports whether the match ended.
func (timer *Timer) Tick(elapsed time.Duration) bool {
	if !timer.Running || elapsed <= 0 {
		return false
	}
	timer.Remaining -= elapsed
	if timer.Remaining > 0 {
		return false
	}
	timer.Remaining, timer.Running, timer.Started = 0, false, false
	return true
}

// Reset ends the match and restores the selected duration.
func (timer *Timer) Reset() {
	timer.Remaining, timer.Running, timer.Started = timer.Steps[timer.Step], false, false
}
