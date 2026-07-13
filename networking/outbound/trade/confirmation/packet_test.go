package confirmation

import "testing"

// TestEncode verifies TRADE_CONFIRMATION encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
