package live

import "testing"

// TestNewHolderStoresInitializedAt verifies holder creation timestamps.
func TestNewHolderStoresInitializedAt(t *testing.T) {
	holder := NewHolder()

	if holder.InitializedAt().IsZero() {
		t.Fatal("expected non-zero initialized time")
	}
}
