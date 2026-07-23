package infobusstart

import "testing"

// TestEncode verifies the Nitro RoomPollDataParser shape.
func TestEncode(t *testing.T) {
	packet, err := Encode("Works?", []string{"No", "Yes"})
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet: %+v %v", packet, err)
	}
}
