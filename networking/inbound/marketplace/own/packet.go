// Package own contains the GET_OWN_MARKETPLACE_OFFERS inbound packet.
package own

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_OWN_MARKETPLACE_OFFERS.
const Header uint16 = 2105

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
