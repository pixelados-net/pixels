package floor

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the complete floor placement payload and strict header.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.String("x"), codec.Int32(3), codec.Int32(4), codec.Int32(6))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PageID != 1 || payload.OfferID != 2 || payload.ExtraData != "x" || payload.X != 3 || payload.Y != 4 || payload.Direction != 6 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header error")
	}
}
