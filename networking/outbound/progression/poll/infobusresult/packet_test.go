package infobusresult

import "testing"

// TestEncode verifies the Nitro RoomPollResultParser shape.
func TestEncode(t *testing.T) {
	packet, err := Encode("Works?", []Choice{{Text: "Yes", Votes: 3}}, 3)
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet: %+v %v", packet, err)
	}
}
