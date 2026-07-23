package perks

import (
	"encoding/binary"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesPerkCollection verifies Nitro's repeated perk shape.
func TestEncodeWritesPerkCollection(t *testing.T) {
	packet, err := Encode([]Entry{{Code: "CAMERA", Error: "", Allowed: true}, {Code: "TRADE", Error: "locked"}})
	if err != nil {
		t.Fatalf("encode perks: %v", err)
	}
	if binary.BigEndian.Uint32(packet.Payload[:4]) != 2 {
		t.Fatalf("unexpected perk count payload %v", packet.Payload)
	}
	values, rest, err := codec.DecodePayload(nil, append(EntryDefinition, EntryDefinition...), packet.Payload[4:])
	if err != nil {
		t.Fatalf("decode perks: %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected remaining perk bytes %d", len(rest))
	}
	if values[0].String != "CAMERA" || !values[2].Boolean || values[3].String != "TRADE" || values[5].Boolean {
		t.Fatalf("unexpected perk values %#v", values)
	}
}

// BenchmarkEncode measures USER_PERKS packet encoding.
func BenchmarkEncode(b *testing.B) {
	entries := []Entry{{Code: "CAMERA", Allowed: true}, {Code: "HEIGHTMAP_EDITOR_BETA", Allowed: true}, {Code: "USE_GUIDE_TOOL"}}
	b.ReportAllocs()
	for b.Loop() {
		packet, err := Encode(entries)
		if err != nil || packet.Header != Header {
			b.Fatalf("unexpected packet=%#v err=%v", packet, err)
		}
	}
}
