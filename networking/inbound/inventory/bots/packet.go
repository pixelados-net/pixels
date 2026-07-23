// Package bots decodes USER_BOTS inventory requests.
package bots

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BOTS.
const Header uint16 = 3848

// Decode validates an empty USER_BOTS request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
