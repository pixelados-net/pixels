package approve

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies name fields and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("Pixel"), codec.Int32(1))
	value, err := Decode(packet)
	if err != nil || value.Name != "Pixel" || value.Type != 1 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
