package extend

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies a placed rentable mutation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Bool(true), codec.Int32(7), codec.Bool(false))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || !payload.Wall || payload.ItemID != 7 || payload.Buyout {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
