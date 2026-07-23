// Package finished encodes LOVELOCK_FURNI_FINISHED responses.
package finished

import "github.com/niflaot/pixels/networking/codec"

// Header identifies LOVELOCK_FURNI_FINISHED.
const Header uint16 = 770

// Definition describes the renderer parser exactly.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Encode creates one sealed lovelock response.
func Encode(itemID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(itemID))
}
