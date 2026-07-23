package offer

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the rentable offer lookup payload.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Bool(false), codec.String("chair"), codec.Bool(true))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Wall || payload.ProductName != "chair" || !payload.Buyout {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
