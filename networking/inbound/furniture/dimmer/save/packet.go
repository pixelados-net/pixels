// Package save decodes a mood-light preset save.
package save

import "github.com/niflaot/pixels/networking/codec"

// Header is the ITEM_DIMMER_SAVE identifier.
const Header uint16 = 1648

// Payload contains one mood-light preset.
type Payload struct {
	// PresetID identifies the one-based slot.
	PresetID int32
	// Type stores one for full-room or two for background-only.
	Type int32
	// Color stores the preset hexadecimal color.
	Color string
	// Brightness stores the preset brightness.
	Brightness int32
	// Apply reports whether the preset becomes active.
	Apply bool
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField}

// Decode decodes a mood-light preset save.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PresetID: values[0].Int32, Type: values[1].Int32, Color: values[2].String, Brightness: values[3].Int32, Apply: values[4].Boolean}, nil
}
