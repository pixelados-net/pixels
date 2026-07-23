package pickup

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates BOT_PICKUP identity normalization and strict headers.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(-7))
	value, err := Decode(packet)
	if err != nil || value != 7 {
		t.Fatalf("value=%d err=%v", value, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
