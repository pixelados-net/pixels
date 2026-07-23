package start

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies breeding fields and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(2), codec.Int32(4), codec.Int32(5))
	value, err := Decode(packet)
	if err != nil || value.State != 2 || value.PetOneID != 4 || value.PetTwoID != 5 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
