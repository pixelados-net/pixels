// Package initmsg contains the MESSENGER_INIT inbound packet.
package initmsg

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_INIT.
const Header uint16 = 2781

// Decode validates a MESSENGER_INIT packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
