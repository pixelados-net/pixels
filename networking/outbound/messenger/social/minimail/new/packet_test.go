package new

import "testing"

// TestEncode verifies the deprecated MiniMail header.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("encode minimail new: %#v, %v", packet, err)
	}
}
