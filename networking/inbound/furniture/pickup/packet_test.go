package pickup

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies PICKUP_FLOOR_ITEM decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(FloorCategory), codec.Int32(9))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload != (Payload{Category: FloorCategory, ItemID: 9}) {
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
