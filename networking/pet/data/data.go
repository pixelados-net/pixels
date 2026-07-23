// Package data encodes protocol values shared by pet packets.
package data

import (
	"strconv"
	"strings"

	"github.com/niflaot/pixels/networking/codec"
)

// CustomPart stores one renderer-specific pet appearance override.
type CustomPart struct {
	// LayerID identifies the overridden renderer layer.
	LayerID int32
	// PartID identifies the replacement part.
	PartID int32
	// PaletteID identifies the replacement palette.
	PaletteID int32
}

// Figure stores one structured Nitro pet figure.
type Figure struct {
	// TypeID identifies the pet species.
	TypeID int32
	// PaletteID identifies the base palette.
	PaletteID int32
	// Color stores the normalized hexadecimal color without a hash.
	Color string
	// BreedID identifies the breed variant.
	BreedID int32
	// CustomParts stores renderer part overrides.
	CustomParts []CustomPart
}

// Pet stores Nitro's reusable inventory pet data.
type Pet struct {
	// ID identifies the durable pet.
	ID int64
	// Name stores the visible pet name.
	Name string
	// Figure stores structured appearance data.
	Figure Figure
	// Level stores the current level.
	Level int32
}

// AppendFigure appends a structured pet figure.
func AppendFigure(dst []byte, figure Figure) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(figure.TypeID), codec.Int32(figure.PaletteID), codec.String(figure.Color),
		codec.Int32(figure.BreedID), codec.Int32(int32(len(figure.CustomParts))))
	if err != nil {
		return nil, err
	}
	for _, part := range figure.CustomParts {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field},
			codec.Int32(part.LayerID), codec.Int32(part.PartID), codec.Int32(part.PaletteID))
		if err != nil {
			return nil, err
		}
	}
	return payload, nil
}

// AppendPet appends Nitro's reusable inventory pet data.
func AppendPet(dst []byte, pet Pet) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(pet.ID)), codec.String(pet.Name))
	if err != nil {
		return nil, err
	}
	payload, err = AppendFigure(payload, pet.Figure)
	if err != nil {
		return nil, err
	}
	return codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(pet.Level))
}

// FigureString returns the renderer-compatible figure string.
func FigureString(figure Figure) string {
	var builder strings.Builder
	builder.Grow(32 + len(figure.CustomParts)*12)
	builder.WriteString(strconv.Itoa(int(figure.TypeID)))
	builder.WriteByte(' ')
	builder.WriteString(strconv.Itoa(int(figure.PaletteID)))
	builder.WriteByte(' ')
	builder.WriteString(figure.Color)
	builder.WriteByte(' ')
	builder.WriteString(strconv.Itoa(len(figure.CustomParts)))
	for _, part := range figure.CustomParts {
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(int(part.LayerID)))
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(int(part.PartID)))
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(int(part.PaletteID)))
	}
	return builder.String()
}
