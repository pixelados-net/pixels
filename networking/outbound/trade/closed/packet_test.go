package closed

import "testing"

// TestEncode verifies TRADE_CLOSED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 1)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
