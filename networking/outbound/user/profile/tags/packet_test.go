package tags

import "testing"

// TestEncode verifies tag list encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(3, []string{"pixel"})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode tags: %#v, %v", packet, err)
	}
}
