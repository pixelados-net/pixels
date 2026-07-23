package compatibility

import (
	"testing"

	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestRegisterInstallsRetiredHandlers verifies all compatibility handlers register.
func TestRegisterInstallsRetiredHandlers(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry)
	if registry.Len() != 7 {
		t.Fatalf("expected seven compatibility handlers, got %d", registry.Len())
	}
}
