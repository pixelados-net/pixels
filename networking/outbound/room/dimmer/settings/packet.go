// Package settings encodes room mood-light presets.
package settings

import (
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	"github.com/niflaot/pixels/networking/codec"
)

// Header is the ITEM_DIMMER_SETTINGS identifier.
const Header uint16 = 2710

// Encode creates an ITEM_DIMMER_SETTINGS packet.
func Encode(presets []roomdecor.Preset) (codec.Packet, error) {
	selected := int32(1)
	for _, preset := range presets {
		if preset.Selected {
			selected = preset.ID
			break
		}
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(len(presets))), codec.Int32(selected))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, preset := range presets {
		presetType := int32(1)
		if preset.BackgroundOnly {
			presetType = 2
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(preset.ID), codec.Int32(presetType), codec.String(preset.Color), codec.Int32(preset.Brightness))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
