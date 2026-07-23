// Package action decodes UNIT_ACTION requests.
package action

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_ACTION.
const Header uint16 = 2456

// Definition describes UNIT_ACTION fields.
var Definition = codec.Definition{codec.Named("actionId", codec.Int32Field)}

// Decode decodes a UNIT_ACTION packet.
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
