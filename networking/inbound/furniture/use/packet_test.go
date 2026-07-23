package use

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the Nitro furniture-use payload.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(42), codec.Int32(1))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 42 || payload.State != 1 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}

// TestDecodeRejectsHeader verifies directional header validation.
func TestDecodeRejectsHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
