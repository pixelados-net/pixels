package popularrooms

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies both popular-room filters and strict validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("busy"), codec.Int32(2))
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decode(packet)
	if err != nil || value.Query != "busy" || value.PageIndex != 2 {
		t.Fatalf("value=%+v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: packet.Payload}); err == nil {
		t.Fatal("expected unexpected-header error")
	}
	if _, err = Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected missing-payload error")
	}
}
