package command

import (
	"context"
	"testing"
)

// TestSenderContextAndPermissionRequirement verifies Brigadier sender propagation.
func TestSenderContextAndPermissionRequirement(t *testing.T) {
	sender := ConsoleSender{}
	ctx := WithSender(context.Background(), sender)
	resolved, found := SenderFrom(ctx)
	if !found || resolved.Kind() != SenderKindConsole {
		t.Fatalf("expected console sender, found=%v sender=%v", found, resolved)
	}
	if !RequiresPermission("plugin.test.use")(ctx) {
		t.Fatal("expected console permission")
	}
	if RequiresPermission("plugin.test.use")(context.Background()) {
		t.Fatal("expected missing sender denial")
	}
}
