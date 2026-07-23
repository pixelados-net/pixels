// Package confirmed encodes LOVELOCK_FURNI_FRIEND_COMFIRMED responses.
package confirmed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies LOVELOCK_FURNI_FRIEND_COMFIRMED.
const Header uint16 = 382

// Definition describes the renderer parser exactly.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Encode creates one partial lovelock confirmation.
func Encode(itemID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(itemID))
}
