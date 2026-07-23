package configuration

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates BOT_CONFIGURATION fields and strict headers.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(-9), codec.Int32(5))
	payload, err := Decode(packet)
	if err != nil || payload.BotID != 9 || payload.SkillID != 5 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
