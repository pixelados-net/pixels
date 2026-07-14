// Package dance decodes UNIT_DANCE requests.
package dance

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_DANCE.
const Header uint16 = 2080

// Definition describes UNIT_DANCE fields.
var Definition = codec.Definition{codec.Named("danceId", codec.Int32Field)}

// Decode decodes a UNIT_DANCE packet.
func Decode(packet codec.Packet) (int32, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
