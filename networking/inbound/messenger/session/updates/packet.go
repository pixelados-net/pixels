// Package updates contains FRIEND_LIST_UPDATE refresh requests.
package updates

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FRIEND_LIST_UPDATE.
const Header uint16 = 1419

// Decode validates a FRIEND_LIST_UPDATE refresh request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
