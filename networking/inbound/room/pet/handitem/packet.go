// Package handitem decodes UNIT_GIVE_HANDITEM_PET requests.
package handitem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_GIVE_HANDITEM_PET.
const Header uint16 = 2768

// Decode decodes the room-local pet unit identifier.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
