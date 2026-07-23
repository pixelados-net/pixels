package data

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendPet verifies field order and custom-part triples.
func TestAppendPet(t *testing.T) {
	pet := Pet{ID: 7, Name: "Pixel", Level: 4, Figure: Figure{TypeID: 1, PaletteID: 2, Color: "AABBCC", BreedID: 3, CustomParts: []CustomPart{{LayerID: 4, PartID: 5, PaletteID: 6}}}}
	payload, err := AppendPet(nil, pet)
	if err != nil {
		t.Fatalf("append pet: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
		codec.StringField, codec.Int32Field, codec.Int32Field,
	}, payload)
	if err != nil || values[0].Int32 != 7 || values[6].Int32 != 1 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
	parts, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, rest)
	if err != nil || parts[0].Int32 != 4 || parts[3].Int32 != 4 || len(rest) != 0 {
		t.Fatalf("parts=%#v rest=%d err=%v", parts, len(rest), err)
	}
}

// TestFigureString verifies the renderer string excludes the inventory-only breed field.
func TestFigureString(t *testing.T) {
	value := FigureString(Figure{TypeID: 0, PaletteID: 1, Color: "FFFFFF", BreedID: 9, CustomParts: []CustomPart{{LayerID: 2, PartID: 3, PaletteID: 4}}})
	if value != "0 1 FFFFFF 1 2 3 4" {
		t.Fatalf("unexpected figure %q", value)
	}
}
