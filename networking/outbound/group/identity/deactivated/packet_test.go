package deactivated

import "testing"

// TestEncodeWritesHeader verifies the protocol header.
func TestEncodeWritesHeader(t *testing.T) {
	packet, err := Encode(7)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
