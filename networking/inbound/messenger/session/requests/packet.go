// Package requests contains GET_FRIEND_REQUESTS.
package requests

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_FRIEND_REQUESTS.
const Header uint16 = 2448

// Decode validates GET_FRIEND_REQUESTS.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
