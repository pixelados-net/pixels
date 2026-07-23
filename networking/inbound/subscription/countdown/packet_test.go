package countdown

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies strict GET_SECONDS_UNTIL decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("2030-04-05 12:30"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decode(packet)
	if err != nil || value != "2030-04-05 12:30" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: packet.Payload}); err == nil {
		t.Fatal("expected unexpected header")
	}
	if _, err = Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected missing payload")
	}
}
