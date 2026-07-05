package ping

import "testing"

// TestNew verifies CLIENT_PING packet metadata.
func TestNew(t *testing.T) {
	packet := New()

	if packet.Header != Header {
		t.Fatalf("expected header %d, got %d", Header, packet.Header)
	}

	if len(packet.Payload) != 0 {
		t.Fatalf("expected empty payload, got %d bytes", len(packet.Payload))
	}
}
