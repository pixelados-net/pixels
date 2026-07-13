package offers

import "testing"

// TestEncode verifies an empty MARKETPLACE_OFFERS response.
func TestEncode(t *testing.T) {
	packet, err := Encode(nil, 0)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
