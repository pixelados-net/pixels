package list

import "testing"

// TestEncode verifies Nitro's empty list and default-category fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(nil, "")
	if err != nil || packet.Header != Header || len(packet.Payload) != 6 {
		t.Fatalf("unexpected packet=%+v err=%v", packet, err)
	}
}
