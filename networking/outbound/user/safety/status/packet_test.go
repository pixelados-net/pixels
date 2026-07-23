package status

import "testing"

// TestEncode verifies supported safety states.
func TestEncode(t *testing.T) {
	packet, err := Encode(Unlocked)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode safety: %#v, %v", packet, err)
	}
	if _, err = Encode(3); err == nil {
		t.Fatal("expected invalid safety state")
	}
}
