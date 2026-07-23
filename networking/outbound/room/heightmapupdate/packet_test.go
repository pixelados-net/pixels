package heightmapupdate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeRoundTripsTiles verifies the tile count and each x/y/value triple encode in order.
func TestEncodeRoundTripsTiles(t *testing.T) {
	tiles := []Tile{{X: 1, Y: 2, Value: 256}, {X: 3, Y: 4, Value: int16(uint16(0x4100))}}

	packet, err := Encode(tiles)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	definition, _ := definitionFor(len(tiles))
	decoded, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
	if decoded[0].Byte != 2 {
		t.Fatalf("unexpected count %d", decoded[0].Byte)
	}
	for index, expected := range tiles {
		base := 1 + index*3
		if int(decoded[base].Byte) != expected.X || int(decoded[base+1].Byte) != expected.Y || int16(decoded[base+2].Uint16) != expected.Value {
			t.Fatalf("unexpected tile %d %#v", index, decoded[base:base+3])
		}
	}
}

// TestEncodeEmptyTiles verifies an empty tile set still encodes a valid zero count.
func TestEncodeEmptyTiles(t *testing.T) {
	packet, err := Encode(nil)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	definition, _ := definitionFor(0)
	decoded, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded[0].Byte != 0 {
		t.Fatalf("unexpected count %d", decoded[0].Byte)
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
}
