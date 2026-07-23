// Package posture decodes UNIT_POSTURE requests.
package posture

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_POSTURE.
const Header uint16 = 2235

// Definition describes UNIT_POSTURE fields.
var Definition = codec.Definition{codec.Named("posture", codec.Int32Field)}

// Decode decodes a UNIT_POSTURE packet.
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
