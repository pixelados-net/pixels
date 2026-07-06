package cancreate

import "testing"

// TestCommandName verifies the command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name")
	}
}
