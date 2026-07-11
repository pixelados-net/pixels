package request

import "testing"

// TestCommandNameIsStable verifies room word filter request identity.
func TestCommandNameIsStable(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}
