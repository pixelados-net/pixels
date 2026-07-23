package completed

import "testing"

// TestEncode verifies TRADE_COMPLETED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
