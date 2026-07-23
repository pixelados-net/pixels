package start

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies strict compatibility packet decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("legacy"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decode(packet)
	if err != nil || value != "legacy" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: packet.Payload}); err == nil {
		t.Fatal("expected unexpected header")
	}
	if _, err = Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected missing payload")
	}
}
