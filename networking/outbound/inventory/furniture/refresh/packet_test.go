package refresh

import "testing"

// TestEncode verifies FURNITURE_INVENTORY_REFRESH encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}
	if len(packet.Payload) != 0 {
		t.Fatalf("expected empty payload, got %d bytes", len(packet.Payload))
	}
}
