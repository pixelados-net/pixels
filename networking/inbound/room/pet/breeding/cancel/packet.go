// Package cancel decodes PET_CANCEL_BREEDING requests.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_CANCEL_BREEDING.
const Header uint16 = 2713

// Decode decodes the breeding nest item identifier.
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
