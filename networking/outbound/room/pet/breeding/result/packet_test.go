package result

import "testing"

// TestEncode verifies both result records are serialized.
func TestEncode(t *testing.T) {
	packet, err := Encode(Result{StuffID: 1, ProductCode: "pet0", UserID: 7, UserName: "demo"}, Result{StuffID: 2, ProductCode: "pet0", UserID: 8, UserName: "alice", HasMutation: true})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
