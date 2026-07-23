package blockedtiles

import (
	"encoding/binary"
	"testing"
)

// TestEncodeWritesBlockedTiles verifies count and coordinate order.
func TestEncodeWritesBlockedTiles(t *testing.T) {
	packet, err := Encode([]Tile{{X: 1, Y: 2}, {X: 3, Y: 4}})
	if err != nil {
		t.Fatalf("encode blocked tiles: %v", err)
	}
	if packet.Header != Header || len(packet.Payload) != 20 {
		t.Fatalf("unexpected packet %#v", packet)
	}
	values := []uint32{
		binary.BigEndian.Uint32(packet.Payload[0:4]), binary.BigEndian.Uint32(packet.Payload[4:8]),
		binary.BigEndian.Uint32(packet.Payload[8:12]), binary.BigEndian.Uint32(packet.Payload[12:16]),
		binary.BigEndian.Uint32(packet.Payload[16:20]),
	}
	if values[0] != 2 || values[1] != 1 || values[2] != 2 || values[3] != 3 || values[4] != 4 {
		t.Fatalf("unexpected values %#v", values)
	}
}
