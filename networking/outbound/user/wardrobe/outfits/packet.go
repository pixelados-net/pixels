// Package outfits contains the USER_OUTFITS outbound packet.
package outfits

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_OUTFITS.
const Header uint16 = 3315

// Definition describes USER_OUTFITS list fields.
var Definition = codec.Definition{codec.Named("pageId", codec.Int32Field), codec.Named("outfitCount", codec.Int32Field)}

// OutfitDefinition describes one wardrobe outfit.
var OutfitDefinition = codec.Definition{codec.Named("slotId", codec.Int32Field), codec.Named("figure", codec.StringField), codec.Named("gender", codec.StringField)}

// Encode creates a USER_OUTFITS packet from parallel bounded slices.
func Encode(pageID int32, slotIDs []int32, figures []string, genders []string) (codec.Packet, error) {
	if len(slotIDs) != len(figures) || len(slotIDs) != len(genders) {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(pageID), codec.Int32(int32(len(slotIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for index, slotID := range slotIDs {
		payload, err = codec.AppendPayload(payload, OutfitDefinition, codec.Int32(slotID), codec.String(figures[index]), codec.String(genders[index]))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
