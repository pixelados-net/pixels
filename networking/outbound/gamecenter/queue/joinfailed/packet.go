// Package joinfailed encodes JOININGQUEUEFAILED responses.
package joinfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies JOININGQUEUEFAILED.
const Header uint16 = 3035

// Encode creates one JOININGQUEUEFAILED response.
func Encode(gameTypeID int32, reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(gameTypeID), codec.Int32(reason))
}
