package songinfo

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Nitro's counted song identifier payload.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(-1))
	ids, err := Decode(packet)
	if err != nil || len(ids) != 1 || ids[0] != -1 {
		t.Fatalf("unexpected ids=%+v err=%v", ids, err)
	}
}
