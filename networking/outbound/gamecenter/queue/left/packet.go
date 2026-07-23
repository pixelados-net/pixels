// Package left encodes LEFTQUEUE responses.
package left

import "github.com/niflaot/pixels/networking/codec"

// Header identifies LEFTQUEUE.
const Header uint16 = 1477

// Encode creates one LEFTQUEUE response.
func Encode(gameTypeID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(gameTypeID))
}
