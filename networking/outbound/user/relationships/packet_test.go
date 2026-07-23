package relationships

import "testing"

// TestEncode verifies MESSENGER_RELATIONSHIPS framing.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, []Entry{{Type: 1, Count: 2, FriendID: 8, FriendName: "demo", FriendLook: "look"}})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode packet=%#v err=%v", packet, err)
	}
}
