package setfavorite

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeReadsPayload verifies strict field decoding.
func TestDecodeReadsPayload(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(7))
	value, err := Decode(packet)
	if err != nil || value != 7 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
