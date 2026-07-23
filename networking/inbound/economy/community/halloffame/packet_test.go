package halloffame

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the goal-code request.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("habboFameComp"))
	value, err := Decode(packet)
	if err != nil || value != "habboFameComp" {
		t.Fatalf("unexpected value=%q err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: packet.Payload}); err == nil {
		t.Fatal("expected unexpected-header error")
	}
	if _, err = Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected missing-payload error")
	}
}
