package skillsave

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates BOT_SKILL_SAVE fields and strict headers.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(-9), codec.Int32(2), codec.String("hello"))
	payload, err := Decode(packet)
	if err != nil || payload.BotID != 9 || payload.SkillID != 2 || payload.Data != "hello" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
}
