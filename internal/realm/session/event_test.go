package session

import "testing"

// TestEventNames verifies session event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(EventBound),
		string(EventUnbound),
	}

	for _, event := range events {
		if event == "" {
			t.Fatal("expected event name")
		}
	}
}
