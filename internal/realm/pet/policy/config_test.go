package policy

import "testing"

// TestNormalize verifies unsafe values receive bounded defaults.
func TestNormalize(t *testing.T) {
	config := (Config{MaxPerRoom: -1, MaxPerOwnerRoom: 99, InventoryFragmentSize: 999}).Normalize()
	if config.MaxPerRoom != 25 || config.MaxPerOwnerRoom != 10 || config.InventoryFragmentSize != 100 {
		t.Fatalf("unexpected config %#v", config)
	}
}
