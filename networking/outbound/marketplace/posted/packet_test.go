package posted

import "testing"

// TestEncode verifies MARKETPLACE_ITEM_POSTED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
