package list

import (
	"errors"
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// TestEncodeRoundTripsFragmentAndItems verifies encoding and decoding a fragment of items.
func TestEncodeRoundTripsFragmentAndItems(t *testing.T) {
	items := []Item{
		{ID: 1, SpriteID: 22, ExtraData: "0", AllowInventoryStack: true},
		{ID: 2, SpriteID: 39, ExtraData: "0", AllowInventoryStack: false},
	}

	packet, err := Encode(1, 2, items)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	values, rest, err := codec.DecodePayload(nil, headerDefinition(), packet.Payload)
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	if values[0].Int32 != 2 || values[1].Int32 != 0 || values[2].Int32 != 2 {
		t.Fatalf("unexpected header values %#v", values)
	}

	for index, expected := range items {
		var itemValues []codec.Value
		itemValues, rest, err = codec.DecodePayload(nil, itemDefinition(), rest)
		if err != nil {
			t.Fatalf("decode item %d: %v", index, err)
		}
		if itemValues[0].Int32 != int32(expected.ID) || itemValues[2].Int32 != int32(expected.ID) {
			t.Fatalf("unexpected item id %#v", itemValues)
		}
		if itemValues[1].String != floorTypeCode {
			t.Fatalf("unexpected type code %#v", itemValues[1])
		}
		if itemValues[3].Int32 != int32(expected.SpriteID) {
			t.Fatalf("unexpected sprite id %#v", itemValues[3])
		}
		if itemValues[6].String != expected.ExtraData {
			t.Fatalf("unexpected extradata %#v", itemValues[6])
		}
		if itemValues[9].Boolean != expected.AllowInventoryStack {
			t.Fatalf("unexpected allow inventory stack %#v", itemValues[9])
		}
		var floorValues []codec.Value
		floorValues, rest, err = codec.DecodePayload(nil, floorDefinition(), rest)
		if err != nil || len(floorValues) != 2 {
			t.Fatalf("decode floor trailer %d: values=%#v error=%v", index, floorValues, err)
		}
	}
	if len(rest) != 0 {
		t.Fatalf("expected fully consumed payload, %d bytes left", len(rest))
	}
}

// TestEncodeSpecializedObjectDataPreservesGroupIdentity verifies inventory items can carry group object data.
func TestEncodeSpecializedObjectDataPreservesGroupIdentity(t *testing.T) {
	packet, err := Encode(1, 1, []Item{{
		ID: 7, SpriteID: 4254,
		Data: stuffdata.StringArray([]string{"0", "2", "b14014s05050", "#ff0000", "#00ff00"}),
	}})
	if err != nil {
		t.Fatalf("encode group item: %v", err)
	}
	_, rest, err := codec.DecodePayload(nil, headerDefinition(), packet.Payload)
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, itemPrefixDefinition(), rest)
	if err != nil || values[3].Int32 != 4254 {
		t.Fatalf("unexpected prefix %#v error=%v", values, err)
	}
	data, rest, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field,
		codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.StringField,
	}, rest)
	if err != nil || data[0].Int32 != 2 || data[1].Int32 != 5 || data[3].String != "2" || data[4].String != "b14014s05050" {
		t.Fatalf("unexpected group data %#v error=%v", data, err)
	}
	_, rest, err = codec.DecodePayload(nil, itemSuffixDefinition(), rest)
	if err != nil {
		t.Fatalf("decode suffix: %v", err)
	}
	_, rest, err = codec.DecodePayload(nil, floorDefinition(), rest)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode floor trailer: rest=%d error=%v", len(rest), err)
	}
}

// TestEncodeWallItemOmitsFloorTrailer verifies Nitro's compact wall inventory shape.
func TestEncodeWallItemOmitsFloorTrailer(t *testing.T) {
	packet, err := Encode(1, 1, []Item{{ID: 3, SpriteID: 3002, Kind: KindWall, Category: CategoryFloor, ExtraData: "501"}})
	if err != nil {
		t.Fatalf("encode wall item: %v", err)
	}
	_, rest, err := codec.DecodePayload(nil, headerDefinition(), packet.Payload)
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, itemDefinition(), rest)
	if err != nil || values[1].String != wallTypeCode || values[4].Int32 != int32(CategoryFloor) || len(rest) != 0 {
		t.Fatalf("unexpected wall values %#v rest=%d error=%v", values, len(rest), err)
	}
}

// TestEncodeGiftUsesPositiveIDAndPackedVariant verifies unopened present inventory semantics.
func TestEncodeGiftUsesPositiveIDAndPackedVariant(t *testing.T) {
	packet, err := Encode(1, 1, []Item{{ID: 40, SpriteID: 3379, GiftWrapped: true, GiftBoxID: 2, GiftRibbonID: 7}})
	if err != nil {
		t.Fatalf("encode gift: %v", err)
	}
	_, rest, err := codec.DecodePayload(nil, headerDefinition(), packet.Payload)
	if err != nil {
		t.Fatalf("decode gift header: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, itemDefinition(), rest)
	if err != nil || values[0].Int32 != 40 || values[3].Int32 != 3379 || values[4].Int32 != 2007 {
		t.Fatalf("unexpected gift values %#v error=%v", values, err)
	}
	floor, rest, err := codec.DecodePayload(nil, floorDefinition(), rest)
	if err != nil || floor[1].Int32 != 2007 || len(rest) != 0 {
		t.Fatalf("unexpected gift trailer %#v rest=%d error=%v", floor, len(rest), err)
	}
}

// TestEncodeRejectsUnsupportedKind verifies invalid discriminators fail explicitly.
func TestEncodeRejectsUnsupportedKind(t *testing.T) {
	_, err := Encode(1, 1, []Item{{ID: 1, Kind: Kind("ceiling")}})
	if !errors.Is(err, ErrUnsupportedKind) {
		t.Fatalf("expected unsupported kind error, got %v", err)
	}
}

// TestEncodeEmptyFragmentProducesZeroCount verifies the empty-inventory shape.
func TestEncodeEmptyFragmentProducesZeroCount(t *testing.T) {
	packet, err := Encode(1, 1, nil)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	values, rest, err := codec.DecodePayload(nil, headerDefinition(), packet.Payload)
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	if values[2].Int32 != 0 || len(rest) != 0 {
		t.Fatalf("expected zero items and empty payload, got %#v rest=%d", values, len(rest))
	}
}

// TestEncodeRejectsOversizedExtraData verifies item encoding errors surface.
func TestEncodeRejectsOversizedExtraData(t *testing.T) {
	oversized := strings.Repeat("x", 1<<16)

	_, err := Encode(1, 1, []Item{{ID: 1, ExtraData: oversized}})
	if err == nil {
		t.Fatal("expected oversized extradata to fail encoding")
	}
}
