package create

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_CREATE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition,
		codec.String("Room"),
		codec.String("Description"),
		codec.String("model_a"),
		codec.Int32(2),
		codec.Int32(25),
		codec.Int32(1),
	)
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.ModelName != "model_a" || payload.MaxVisitors != 25 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
