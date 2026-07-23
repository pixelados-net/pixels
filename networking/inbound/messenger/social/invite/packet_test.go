package invite

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRecipientsAndMessage verifies SEND_ROOM_INVITE decoding.
func TestDecodeReadsRecipientsAndMessage(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField}, codec.Int32(1), codec.Int32(7), codec.String("join"))
	payload, err := Decode(packet)
	if err != nil || payload.Message != "join" || payload.PlayerIDs[0] != 7 {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
