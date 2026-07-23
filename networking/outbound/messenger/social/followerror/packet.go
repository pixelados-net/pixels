// Package followerror contains MESSENGER_FOLLOW_FAILED.
package followerror

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_FOLLOW_FAILED.
const Header uint16 = 3048

// Encode creates MESSENGER_FOLLOW_FAILED.
func Encode(errorCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(errorCode))
}
