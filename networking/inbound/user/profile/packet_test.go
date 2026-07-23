package profile

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsProfileRequest verifies USER_PROFILE decoding.
func TestDecodeReadsProfileRequest(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.BooleanField}, codec.Int32(7), codec.Bool(true))
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || !payload.OpenWindow {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
