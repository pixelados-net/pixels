package owner

import "testing"

// TestEncodeWritesEmptyOwnerPacket verifies owner packet encoding.
func TestEncodeWritesEmptyOwnerPacket(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("unexpected packet %#v err=%v", packet, err)
	}
}
