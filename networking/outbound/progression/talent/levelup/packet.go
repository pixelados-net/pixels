// Package levelup encodes TALENT_TRACK_LEVEL_UP responses.
package levelup

import (
	"errors"

	"github.com/niflaot/pixels/networking/codec"
	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
)

// Header identifies TALENT_TRACK_LEVEL_UP.
const Header uint16 = 638

// ErrMixedRewards reports Nitro's parser bug for mixed reward families.
var ErrMixedRewards = errors.New("talent level-up cannot mix perks and products")

// Encode creates one renderer-safe talent level-up response.
func Encode(name string, level int32, perkIDs []int32, products []talentdata.Product) (codec.Packet, error) {
	if len(perkIDs) > 0 && len(products) > 0 {
		return codec.Packet{}, ErrMixedRewards
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field}, codec.String(name), codec.Int32(level), codec.Int32(int32(len(perkIDs))))
	for _, perkID := range perkIDs {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(perkID))
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(products))))
	}
	for _, product := range products {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(product.Name), codec.Int32(product.Value))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
