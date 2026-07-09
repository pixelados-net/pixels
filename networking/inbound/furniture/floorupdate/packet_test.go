package floorupdate

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies MOVE_FLOOR_ITEM decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(9), codec.Int32(3), codec.Int32(4), codec.Int32(2))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload != (Payload{ItemID: 9, X: 3, Y: 4, Rotation: 2}) {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeRejectsUnexpectedHeader verifies header validation.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
