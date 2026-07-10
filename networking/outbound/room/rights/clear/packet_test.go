package clear

import "testing"

// TestEncodeWritesEmptyClearPacket verifies clear packet encoding.
func TestEncodeWritesEmptyClearPacket(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("unexpected packet %#v err=%v", packet, err)
	}
}
