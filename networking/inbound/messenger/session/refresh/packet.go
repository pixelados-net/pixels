// Package refresh contains the MESSENGER_FRIENDS refresh inbound packet.
package refresh

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_FRIENDS.
const Header uint16 = 1523

// Decode validates a MESSENGER_FRIENDS packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
