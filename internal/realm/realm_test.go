package realm

import "testing"

// TestRealmReady verifies named realms can accept sessions.
func TestRealmReady(t *testing.T) {
	realm := New("local")

	if !realm.Ready() {
		t.Fatal("expected named realm to be ready")
	}
}

// TestRealmReadyRequiresName verifies empty realms are not ready.
func TestRealmReadyRequiresName(t *testing.T) {
	realm := New("")

	if realm.Ready() {
		t.Fatal("expected empty realm to be unavailable")
	}
}
