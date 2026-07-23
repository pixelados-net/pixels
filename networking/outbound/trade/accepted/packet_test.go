package accepted

import "testing"

// TestEncode verifies TRADE_ACCEPTED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, true)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
