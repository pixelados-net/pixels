package competitive

import "testing"

// TestEncode verifies the leaderboard metadata response.
func TestEncode(t *testing.T) {
	packet, err := Encode(2026, 30, 0, 0, 900)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode failed: %v", err)
	}
}
