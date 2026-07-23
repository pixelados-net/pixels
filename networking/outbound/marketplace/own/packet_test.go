package own

import "testing"

// TestEncode verifies an empty OWN_MARKETPLACE_OFFERS response.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, nil)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
