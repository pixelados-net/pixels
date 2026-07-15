package give

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies hand-item recipient fields and header validation.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9))
	payload, err := Decode(packet)
	if err != nil || payload.UnitID != 9 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
