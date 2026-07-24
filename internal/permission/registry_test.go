package permission

import (
	"errors"
	"testing"
)

// TestCMSPermissionNodesAreRegistered verifies website capabilities can be assigned through the API.
func TestCMSPermissionNodesAreRegistered(t *testing.T) {
	nodes := []Node{
		CMSNewsManage,
		CMSMaintenanceManage,
		CMSMaintenanceEarlyAccessManage,
		CMSPermissionGroupsView,
		CMSPermissionGroupsCreate,
		CMSPermissionGroupsUpdate,
		CMSPermissionGroupNodesManage,
	}
	for _, node := range nodes {
		if !Registered(node) {
			t.Fatalf("expected CMS node %q to be registered", node)
		}
	}
}

// TestRegisterPluginNodeAddsDescribedDynamicNode verifies runtime permission declarations.
func TestRegisterPluginNodeAddsDescribedDynamicNode(t *testing.T) {
	node := Node("plugin.registry-test.hello.use")
	if err := RegisterPluginNode(node, "Use hello command"); err != nil {
		t.Fatalf("register plugin node: %v", err)
	}
	if !Registered(node) {
		t.Fatal("expected dynamic node registration")
	}
	if err := RegisterPluginNode(node, "duplicate"); !errors.Is(err, ErrDuplicateRegistration) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
	if err := RegisterPluginNode("players.invalid", "wrong namespace"); !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("expected namespace rejection, got %v", err)
	}
}

// TestRegisterNodeExposesMetadata verifies stable node catalog behavior.
func TestRegisterNodeExposesMetadata(t *testing.T) {
	first := RegisterNode("test.registry.alpha", "TEST_ALPHA")
	second := RegisterNode("test.registry.beta", "")
	if first == second || !Registered(first) || !Registered(second) {
		t.Fatal("expected distinct registered nodes")
	}

	found := false
	for _, registration := range RegisteredNodes() {
		if registration.Node == first {
			found = registration.PerkName == "TEST_ALPHA" && registration.Package != ""
		}
	}
	if !found {
		t.Fatal("expected registered metadata")
	}
}

// TestRegisterNodeRejectsDuplicates verifies collisions fail during initialization.
func TestRegisterNodeRejectsDuplicates(t *testing.T) {
	node := RegisterNode("test.registry.duplicate", "")
	deferred := false
	func() {
		defer func() {
			deferred = recover() != nil
		}()
		RegisterNode(node, "")
	}()
	if !deferred {
		t.Fatal("expected duplicate registration panic")
	}
}

// TestRegisterNodeRejectsInvalidSyntax verifies invalid declarations fail at boot.
func TestRegisterNodeRejectsInvalidSyntax(t *testing.T) {
	deferred := false
	func() {
		defer func() {
			deferred = recover() != nil
		}()
		RegisterNode("bad..node", "")
	}()
	if !deferred {
		t.Fatal("expected invalid registration panic")
	}
}

// BenchmarkRegisteredNodeLookup measures registry read cost.
func BenchmarkRegisteredNodeLookup(b *testing.B) {
	node := RegisterNode("benchmark.registry.lookup", "")
	b.ReportAllocs()
	for b.Loop() {
		if !Registered(node) {
			b.Fatal("expected registered node")
		}
	}
}
