// Package products contains the CRAFTABLE_PRODUCTS outbound packet.
package products

import (
	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies CRAFTABLE_PRODUCTS.
const Header uint16 = 1000

// Encode creates one visible recipe and aggregate ingredient catalog packet.
func Encode(recipes []craftrecord.Recipe) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(recipes))))
	if err != nil {
		return codec.Packet{}, err
	}
	ingredients := make([]string, 0)
	seen := make(map[string]struct{})
	for _, recipe := range recipes {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField}, codec.String(recipe.Name), codec.String(recipe.RewardName))
		if err != nil {
			return codec.Packet{}, err
		}
		for _, ingredient := range recipe.Ingredients {
			if _, found := seen[ingredient.Name]; found {
				continue
			}
			seen[ingredient.Name] = struct{}{}
			ingredients = append(ingredients, ingredient.Name)
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(ingredients))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, name := range ingredients {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(name))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
