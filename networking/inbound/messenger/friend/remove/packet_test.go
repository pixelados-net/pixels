package remove

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsFriendIDs verifies REMOVE_FRIEND decoding.
func TestDecodeReadsFriendIDs(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(7))
	ids, err := Decode(packet)
	if err != nil || len(ids) != 1 || ids[0] != 7 {
		t.Fatalf("unexpected ids=%#v err=%v", ids, err)
	}
}
