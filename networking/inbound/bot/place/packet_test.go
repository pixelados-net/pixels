package place

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates BOT_PLACE fields and strict headers.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(2), codec.Int32(3))
	payload, err := Decode(packet)
	if err != nil || payload.BotID != 7 || payload.X != 2 || payload.Y != 3 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
