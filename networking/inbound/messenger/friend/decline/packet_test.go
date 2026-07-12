package decline

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsModeAndIDs verifies DECLINE_FRIEND decoding.
func TestDecodeReadsModeAndIDs(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field}, codec.Bool(false), codec.Int32(1), codec.Int32(7))
	payload, err := Decode(packet)
	if err != nil || payload.All || len(payload.PlayerIDs) != 1 {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
