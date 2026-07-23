package privatechat

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsTargetAndMessage verifies MESSENGER_CHAT decoding.
func TestDecodeReadsTargetAndMessage(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(7), codec.String("hello"))
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || payload.Message != "hello" {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
