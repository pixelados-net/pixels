// Package limited contains the GET_LIMITED_OFFER_APPEARING_NEXT inbound packet.
package limited

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_LIMITED_OFFER_APPEARING_NEXT packet identifier.
	Header uint16 = 410
)

// Decode validates a GET_LIMITED_OFFER_APPEARING_NEXT packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
