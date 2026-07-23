// Package track encodes HELPER_TALENT_TRACK responses.
package track

import (
	"github.com/niflaot/pixels/networking/codec"
	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
)

// Header identifies HELPER_TALENT_TRACK.
const Header uint16 = 3406

// Encode creates one complete nested talent track response.
func Encode(name string, levels []talentdata.Level) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(name), codec.Int32(int32(len(levels))))
	for _, level := range levels {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(level.ID), codec.Int32(level.State), codec.Int32(int32(len(level.Tasks))))
		for _, task := range level.Tasks {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(task.ID), codec.Int32(task.Index), codec.String(task.BadgeCode), codec.Int32(task.State), codec.Int32(task.Progress), codec.Int32(task.RequiredProgress))
		}
		if err == nil {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(level.Perks))))
		}
		for _, perk := range level.Perks {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(perk))
		}
		if err == nil {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(level.Products))))
		}
		for _, product := range level.Products {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(product.Name), codec.Int32(product.Value))
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
