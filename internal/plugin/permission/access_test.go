package permission

import (
	"errors"
	"testing"

	permissiondomain "github.com/niflaot/pixels/internal/permission"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
)

// TestAccessRegistersOnlyLocalNamespacedNodes verifies permission ownership.
func TestAccessRegistersOnlyLocalNamespacedNodes(t *testing.T) {
	access := NewAccess(pluginruntime.NewScope("permission-access-test"))
	if err := access.Register("hello.use", "Use hello"); err != nil {
		t.Fatalf("register permission: %v", err)
	}
	if !permissiondomain.Registered("plugin.permission-access-test.hello.use") {
		t.Fatal("expected full plugin permission registration")
	}
	if err := access.Register("plugin.other.node", "invalid"); !errors.Is(err, pluginruntime.ErrInvalidPermission) {
		t.Fatalf("expected local node rejection, got %v", err)
	}
	if err := access.Register("", "invalid"); !errors.Is(err, pluginruntime.ErrInvalidPermission) {
		t.Fatalf("expected empty node rejection, got %v", err)
	}
}
