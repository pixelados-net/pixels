package unseen

import (
	"errors"
	"math"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeOwned verifies UNSEEN_ITEMS owned furniture encoding.
func TestEncodeOwned(t *testing.T) {
	packet, err := EncodeOwned([]int64{42, 43})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	})
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 1 || values[1].Int32 != ownedFurnitureCategory ||
		values[2].Int32 != 2 || values[3].Int32 != 42 || values[4].Int32 != 43 {
		t.Fatalf("unexpected values %#v", values)
	}
}

// TestEncodeOwnedRejectsProtocolOverflow verifies ids cannot truncate on the wire.
func TestEncodeOwnedRejectsProtocolOverflow(t *testing.T) {
	_, err := EncodeOwned([]int64{math.MaxInt64})
	if !errors.Is(err, ErrItemIDRange) {
		t.Fatalf("expected protocol range error, got %v", err)
	}
}

// TestEncodeOwnedAllowsEmptyCategory verifies a valid empty unseen update.
func TestEncodeOwnedAllowsEmptyCategory(t *testing.T) {
	packet, err := EncodeOwned(nil)
	if err != nil || len(packet.Payload) != 12 {
		t.Fatalf("unexpected packet %#v error %v", packet, err)
	}
}
