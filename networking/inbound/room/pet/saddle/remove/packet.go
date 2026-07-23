// Package remove decodes REMOVE_PET_SADDLE requests.
package remove

import "github.com/niflaot/pixels/networking/codec"

// Header identifies REMOVE_PET_SADDLE.
const Header uint16 = 186

// Decode decodes the pet identifier.
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
