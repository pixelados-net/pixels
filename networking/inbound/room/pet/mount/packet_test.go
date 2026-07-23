package mount

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies mount fields and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(5), codec.Bool(true))
	value, err := Decode(packet)
	if err != nil || value.PetID != 5 || !value.Mount {
		t.Fatalf("value=%#v err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
