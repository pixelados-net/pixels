package openfailed

import "testing"

// TestEncode verifies TRADE_OPEN_FAILED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, "target")
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
