package wall

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the complete wall placement payload and strict header.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.String("x"), codec.String(":w=1,2 l=3,4 r"))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PageID != 1 || payload.OfferID != 2 || payload.ExtraData != "x" || payload.WallPosition != ":w=1,2 l=3,4 r" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header error")
	}
}
