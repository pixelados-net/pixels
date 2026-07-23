// Package bus contains local event publishing primitives.
package bus

import (
	"errors"
	"time"
)

var (
	// ErrInvalidEvent reports an event without a name.
	ErrInvalidEvent = errors.New("invalid event")

	// ErrInvalidHandler reports a missing event handler.
	ErrInvalidHandler = errors.New("invalid event handler")
)

// Name identifies an event topic.
type Name string

const (
	// PriorityHigh runs before normal and low priority listeners.
	PriorityHigh = 100

	// PriorityNormal is the default listener priority.
	PriorityNormal = 0

	// PriorityLow runs after normal and high priority listeners.
	PriorityLow = -100
)

// Event describes something that already happened.
type Event struct {
	// Name identifies the event topic.
	Name Name

	// At stores when the event happened.
	At time.Time

	// Payload stores the event-specific value.
	Payload any
}

// Valid reports whether the event can be published.
func (event Event) Valid() bool {
	return event.Name != ""
}

// WithTime returns the event with a default timestamp when missing.
func (event Event) WithTime(now time.Time) Event {
	if event.At.IsZero() {
		event.At = now
	}

	return event
}
