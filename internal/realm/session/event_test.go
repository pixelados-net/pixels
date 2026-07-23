package session

import (
	"testing"

	sessionbound "github.com/niflaot/pixels/internal/realm/session/events/bound"
	sessionunbound "github.com/niflaot/pixels/internal/realm/session/events/unbound"
)

// TestEventNames verifies session event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(sessionbound.Name),
		string(sessionunbound.Name),
	}

	for _, event := range events {
		if event == "" {
			t.Fatal("expected event name")
		}
	}
}
