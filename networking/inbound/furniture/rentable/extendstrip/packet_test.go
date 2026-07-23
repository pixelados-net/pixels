package extendstrip

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies an inventory rentable mutation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Bool(true))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 7 || !payload.Buyout {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
