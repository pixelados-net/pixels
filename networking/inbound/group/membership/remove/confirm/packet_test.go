package removeconfirm

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeReadsPayload verifies strict field decoding.
func TestDecodeReadsPayload(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(8))
	value, err := Decode(packet)
	if err != nil || value.GroupID != 7 || value.PlayerID != 8 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
