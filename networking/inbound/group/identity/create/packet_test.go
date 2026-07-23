package create

import (
	"encoding/base64"
	"errors"
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsNitroCreatorPayload verifies the exact badge value count emitted by Nitro.
func TestDecodeReadsNitroCreatorPayload(t *testing.T) {
	payloadBytes, err := base64.StdEncoding.DecodeString("AAlFbCBwcm9tYXgAHUVsIGdydXBvIG1hcyBwcm9tYXggZGVsIG11bmRvAAAAggAAABgAAAAXAAAACQAAAAEAAAAKAAAABAAAAAIAAAABAAAABwAAAAsAAAADAAAABA==")
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	packet := codec.Packet{Header: Header, Payload: payloadBytes}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Name != "El promax" || payload.Description != "El grupo mas promax del mundo" || payload.RoomID != 130 || payload.ColorA != 24 || payload.ColorB != 23 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	expected := []grouprecord.BadgePart{
		{Ordinal: 0, Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 10, Position: 4},
		{Ordinal: 1, Kind: grouprecord.BadgeSymbol, ElementID: 2, ColorID: 1, Position: 7},
		{Ordinal: 2, Kind: grouprecord.BadgeSymbol, ElementID: 11, ColorID: 3, Position: 4},
	}
	if len(payload.Parts) != len(expected) {
		t.Fatalf("parts=%d want %d", len(payload.Parts), len(expected))
	}
	for index, part := range expected {
		if payload.Parts[index] != part {
			t.Fatalf("part %d=%#v want %#v", index, payload.Parts[index], part)
		}
	}
}

// TestDecodeRejectsInvalidBadgeValueCounts verifies triple alignment and the five-layer bound.
func TestDecodeRejectsInvalidBadgeValueCounts(t *testing.T) {
	tests := []struct {
		name  string
		count int32
	}{
		{name: "empty", count: 0},
		{name: "partial triple", count: 2},
		{name: "unaligned", count: 4},
		{name: "over five parts", count: 18},
	}
	definition := codec.Definition{codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet, err := codec.NewPacket(Header, definition, codec.String("Pixels"), codec.String("Group"), codec.Int32(3), codec.Int32(1), codec.Int32(2), codec.Int32(test.count))
			if err != nil {
				t.Fatalf("new packet: %v", err)
			}
			_, err = Decode(packet)
			if !errors.Is(err, codec.ErrInvalidField) {
				t.Fatalf("err=%v want %v", err, codec.ErrInvalidField)
			}
		})
	}
}
