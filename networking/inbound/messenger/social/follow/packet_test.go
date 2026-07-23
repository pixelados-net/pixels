package follow

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsPlayerID verifies FOLLOW_FRIEND decoding.
func TestDecodeReadsPlayerID(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(7))
	playerID, err := Decode(packet)
	if err != nil || playerID != 7 {
		t.Fatalf("unexpected player=%d err=%v", playerID, err)
	}
}
