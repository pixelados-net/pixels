// Package breeds encodes CATALOG_RECEIVE_PET_BREEDS.
package breeds

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CATALOG_RECEIVE_PET_BREEDS.
const Header uint16 = 3331

// Palette stores one sellable pet palette record.
type Palette struct {
	// TypeID identifies the pet species.
	TypeID int32
	// BreedID identifies the breed.
	BreedID int32
	// PaletteID identifies the renderer palette.
	PaletteID int32
	// Sellable reports whether clients may buy the variant.
	Sellable bool
	// Rare reports whether the variant is rare.
	Rare bool
}

// Encode creates CATALOG_RECEIVE_PET_BREEDS.
func Encode(productCode string, palettes []Palette) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(productCode), codec.Int32(int32(len(palettes))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, palette := range palettes {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.BooleanField},
			codec.Int32(palette.TypeID), codec.Int32(palette.BreedID), codec.Int32(palette.PaletteID), codec.Bool(palette.Sellable), codec.Bool(palette.Rare))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
