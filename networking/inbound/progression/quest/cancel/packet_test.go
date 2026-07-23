package cancel

import "testing"

// TestHeader verifies the packet keeps its registered non-zero header.
func TestHeader(t *testing.T) {
	if Header == 0 {
		t.Fatal("packet header must be non-zero")
	}
}
