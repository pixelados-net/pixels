package flooritems

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeRoundTripsMultipleOwnersAndItems verifies encoding and decoding several records.
func TestEncodeRoundTripsMultipleOwnersAndItems(t *testing.T) {
	owners := []Owner{
		{ID: 1, Name: "demo"},
		{ID: 2, Name: "alice"},
	}
	items := []FloorItem{
		{ID: 1, SpriteID: 22, X: 4, Y: 4, Rotation: 0, Z: "0", ExtraHeight: "", ExtraData: "0", UsagePolicy: 1, OwnerID: 1},
		{ID: 2, SpriteID: 39, X: 3, Y: 3, Rotation: 4, Z: "0", ExtraHeight: "1", ExtraData: "0", UsagePolicy: 1, OwnerID: 1},
		{ID: 3, SpriteID: 46, X: 5, Y: 6, Rotation: 0, Z: "0", ExtraHeight: "", ExtraData: "0", UsagePolicy: 0, OwnerID: 2},
	}

	packet, err := Encode(owners, items)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	rest := packet.Payload
	ownerCount, rest := decodeInt32Count(t, rest)
	if ownerCount != len(owners) {
		t.Fatalf("unexpected owner count %d", ownerCount)
	}
	for index, expected := range owners {
		var owner Owner
		owner, rest = decodeOwner(t, rest)
		if owner != expected {
			t.Fatalf("unexpected owner %d: %#v", index, owner)
		}
	}

	itemCount, rest := decodeInt32Count(t, rest)
	if itemCount != len(items) {
		t.Fatalf("unexpected item count %d", itemCount)
	}
	for index, expected := range items {
		var item FloorItem
		item, rest = decodeFloorItem(t, rest)
		if item != expected {
			t.Fatalf("unexpected item %d: %#v", index, item)
		}
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
}

// TestEncodeRejectsOversizedOwnerName verifies owner encoding errors surface.
func TestEncodeRejectsOversizedOwnerName(t *testing.T) {
	oversized := strings.Repeat("x", 1<<16)

	_, err := Encode([]Owner{{ID: 1, Name: oversized}}, nil)
	if err == nil {
		t.Fatal("expected oversized owner name to fail encoding")
	}
}

// TestEncodeRejectsOversizedItemField verifies floor item encoding errors surface.
func TestEncodeRejectsOversizedItemField(t *testing.T) {
	oversized := strings.Repeat("x", 1<<16)

	_, err := Encode(nil, []FloorItem{{ID: 1, ExtraData: oversized}})
	if err == nil {
		t.Fatal("expected oversized item field to fail encoding")
	}
}

// TestEncodeEmptyProducesZeroCounts verifies the empty-room shape.
func TestEncodeEmptyProducesZeroCounts(t *testing.T) {
	packet, err := Encode(nil, nil)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	ownerCount, rest := decodeInt32Count(t, packet.Payload)
	itemCount, rest := decodeInt32Count(t, rest)
	if ownerCount != 0 || itemCount != 0 || len(rest) != 0 {
		t.Fatalf("expected zero counts and empty payload, owners=%d items=%d rest=%d", ownerCount, itemCount, len(rest))
	}
}

// decodeInt32Count decodes a leading int32 count.
func decodeInt32Count(t *testing.T, src []byte) (int, []byte) {
	t.Helper()

	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, src)
	if err != nil {
		t.Fatalf("decode count: %v", err)
	}

	return int(values[0].Int32), rest
}

// decodeOwner decodes one owner record.
func decodeOwner(t *testing.T, src []byte) (Owner, []byte) {
	t.Helper()

	values, rest, err := codec.DecodePayload(nil, ownerDefinition(), src)
	if err != nil {
		t.Fatalf("decode owner: %v", err)
	}

	return Owner{ID: int64(values[0].Int32), Name: values[1].String}, rest
}

// decodeFloorItem decodes one floor item record.
func decodeFloorItem(t *testing.T, src []byte) (FloorItem, []byte) {
	t.Helper()

	values, rest, err := codec.DecodePayload(nil, floorItemDefinition(), src)
	if err != nil {
		t.Fatalf("decode floor item: %v", err)
	}

	item := FloorItem{
		ID:          int64(values[0].Int32),
		SpriteID:    int(values[1].Int32),
		X:           int(values[2].Int32),
		Y:           int(values[3].Int32),
		Rotation:    int(values[4].Int32),
		Z:           values[5].String,
		ExtraHeight: values[6].String,
		ExtraData:   values[9].String,
		UsagePolicy: values[11].Int32,
		OwnerID:     int64(values[12].Int32),
	}
	if values[7].Int32 != defaultKind || values[8].Int32 != nonLimitedFlag || values[10].Int32 != unknownExpiration {
		t.Fatalf("unexpected constant fields %#v", values)
	}

	return item, rest
}
