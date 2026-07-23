// Package training decodes GET_PET_TRAINING_PANEL requests.
package training

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_PET_TRAINING_PANEL.
const Header uint16 = 2161

// Decode decodes the requested pet identifier.
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
