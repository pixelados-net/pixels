package tokens

import "testing"

// TestEncode verifies MARKETPLACE_TOKENS encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, 5)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
