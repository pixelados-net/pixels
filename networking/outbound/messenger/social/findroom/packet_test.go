package findroom

import "testing"

// TestEncodeWritesSuccess verifies native find-room response creation.
func TestEncodeWritesSuccess(t *testing.T) {
	packet, err := Encode(true)
	if err != nil || packet.Header != Header || len(packet.Payload) != 1 || packet.Payload[0] != 1 {
		t.Fatalf("unexpected packet=%#v err=%v", packet, err)
	}
}
