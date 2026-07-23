package change

import "testing"

// TestEncode verifies name-change result encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, "Pixel", []string{"Pixel2"})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode name change: %#v, %v", packet, err)
	}
}
