// Package get contains the GET_GIFT inbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_GIFT packet identifier.
	Header uint16 = 2436
)

// Decode validates a GET_GIFT packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
