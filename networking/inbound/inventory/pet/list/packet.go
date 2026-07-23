// Package list decodes USER_PETS requests.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_PETS.
const Header uint16 = 3095

// Decode validates an empty USER_PETS request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
