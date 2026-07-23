// Package globalid contains the CONVERT_GLOBAL_ROOM_ID inbound packet.
package globalid

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CONVERT_GLOBAL_ROOM_ID.
const Header uint16 = 314

// Definition describes the renderer-confirmed global identifier string.
var Definition = codec.Definition{codec.Named("globalId", codec.StringField)}

// Payload contains one bounded external room identifier.
type Payload struct {
	// GlobalID stores the identifier exactly as supplied by the link bridge.
	GlobalID string
}

// Decode decodes CONVERT_GLOBAL_ROOM_ID.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	if len(values[0].String) == 0 || len(values[0].String) > 128 {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{GlobalID: values[0].String}, nil
}
