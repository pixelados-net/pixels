package entered

import "testing"

// TestEncode verifies ROOM_ENTER packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("unexpected packet %#v", packet)
	}
}
