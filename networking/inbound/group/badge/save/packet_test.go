package save

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsBoundedBadge verifies layer mapping and count validation.
func TestDecodeReadsBoundedBadge(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(3), codec.Int32(2), codec.Int32(3), codec.Int32(4))
	payload, err := Decode(packet)
	if err != nil || payload.GroupID != 7 || len(payload.Parts) != 1 || payload.Parts[0].ElementID != 2 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}

// TestDecodeRejectsUnalignedBadgeValues verifies that Nitro's integer count contains complete triples.
func TestDecodeRejectsUnalignedBadgeValues(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(4))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	_, err = Decode(packet)
	if !errors.Is(err, codec.ErrInvalidField) {
		t.Fatalf("err=%v want %v", err, codec.ErrInvalidField)
	}
}
