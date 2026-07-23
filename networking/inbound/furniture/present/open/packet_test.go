package open

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies OPEN_PRESENT decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(41))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.ItemID != 41 {
		t.Fatalf("expected item 41, got %d", payload.ItemID)
	}
}

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
