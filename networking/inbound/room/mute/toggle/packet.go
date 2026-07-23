// Package toggle contains the ROOM_MUTE inbound packet.
package toggle

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_MUTE.
	Header uint16 = 3637
)

// Decode validates an empty ROOM_MUTE packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}

	return nil
}
