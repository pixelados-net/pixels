// Package sign decodes UNIT_SIGN requests.
package sign

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_SIGN.
const Header uint16 = 1975

// Definition describes UNIT_SIGN fields.
var Definition = codec.Definition{codec.Named("signId", codec.Int32Field)}

// Decode decodes a UNIT_SIGN packet.
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
