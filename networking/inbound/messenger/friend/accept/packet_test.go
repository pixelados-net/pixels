package accept

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsBoundedIDs verifies ACCEPT_FRIEND decoding.
func TestDecodeReadsBoundedIDs(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(2), codec.Int32(7), codec.Int32(8))
	ids, err := Decode(packet)
	if err != nil || len(ids) != 2 || ids[1] != 8 {
		t.Fatalf("unexpected ids=%#v err=%v", ids, err)
	}
}
