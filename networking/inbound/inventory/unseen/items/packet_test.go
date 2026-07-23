package items

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsItems verifies Nitro's category, count, and item identifiers.
func TestDecodeReadsItems(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(1), codec.Int32(2), codec.Int32(910128), codec.Int32(910129))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Category != 1 || len(payload.ItemIDs) != 2 || payload.ItemIDs[1] != 910129 {
		t.Fatalf("payload=%#v error=%v", payload, err)
	}
}

// TestDecodeRejectsInvalidPackets verifies header, count, and trailing payload validation.
func TestDecodeRejectsInvalidPackets(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
	malformed, err := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(2), codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(malformed); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}
