package buyresult

import "testing"

// TestEncode verifies MARKETPLACE_AFTER_ORDER_STATUS encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(3, 7, 101, 6)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
