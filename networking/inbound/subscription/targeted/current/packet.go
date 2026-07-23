// Package current contains the GET_TARGETED_OFFER inbound packet.
package current

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_TARGETED_OFFER packet identifier.
	Header uint16 = 2487
)

// Decode validates a GET_TARGETED_OFFER packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
