package cancel

import "testing"

// TestEncode verifies MARKETPLACE_CANCEL_RESULT encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, true)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
