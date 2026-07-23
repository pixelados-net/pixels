package settings

import "testing"

// TestCommandNameIsStable verifies room word filter request identity.
func TestCommandNameIsStable(t *testing.T) {
	if (FilterRequestCommand{}).CommandName() != FilterRequestName {
		t.Fatalf("unexpected command name %s", (FilterRequestCommand{}).CommandName())
	}
}
