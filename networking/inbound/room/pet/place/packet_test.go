package place

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies placement fields and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9), codec.Int32(3), codec.Int32(4))
	value, err := Decode(packet)
	if err != nil || value.PetID != 9 || value.X != 3 || value.Y != 4 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
