package remove

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the identifier and strict header.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(5))
	value, err := Decode(packet)
	if err != nil || value != 5 {
		t.Fatalf("value=%d err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
