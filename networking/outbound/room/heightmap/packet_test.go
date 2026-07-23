package heightmap

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeRoundTripsTileValues verifies width, count, and tile values encode in order.
func TestEncodeRoundTripsTileValues(t *testing.T) {
	values := []int16{0, 256, -1, int16(uint16(0x4100))}

	packet, err := Encode(2, values)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	definition, _ := definitionFor(len(values))
	decoded, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
	if decoded[0].Int32 != 2 || decoded[1].Int32 != int32(len(values)) {
		t.Fatalf("unexpected header %#v", decoded[:2])
	}
	for index, expected := range values {
		if int16(decoded[2+index].Uint16) != expected {
			t.Fatalf("unexpected tile %d value %d, want %d", index, int16(decoded[2+index].Uint16), expected)
		}
	}
}

// TestEncodeEmptyTiles verifies an empty tile set still encodes a valid header.
func TestEncodeEmptyTiles(t *testing.T) {
	packet, err := Encode(0, nil)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	definition, _ := definitionFor(0)
	decoded, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded[0].Int32 != 0 || decoded[1].Int32 != 0 {
		t.Fatalf("unexpected header %#v", decoded)
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
}
