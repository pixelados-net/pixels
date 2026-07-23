// Package get contains the CRAFTING_RECIPE outbound packet.
package get

import (
	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies CRAFTING_RECIPE.
const Header uint16 = 2774

// Encode creates one exact ingredient list packet.
func Encode(ingredients []craftrecord.Ingredient) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(ingredients))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, ingredient := range ingredients {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(ingredient.Amount), codec.String(ingredient.Name))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
