package error

import "testing"

// TestEncode verifies the renderer wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header: %d", packet.Header)
	}
	if len(packet.Payload) != 0 {
		t.Fatalf("unexpected payload: %v", packet.Payload)
	}
}
