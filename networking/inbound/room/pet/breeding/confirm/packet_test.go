package confirm

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies confirmation fields and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(3), codec.String("Pixel"), codec.Int32(4), codec.Int32(5))
	value, err := Decode(packet)
	if err != nil || value.NestItemID != 3 || value.Name != "Pixel" || value.PetTwoID != 5 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
