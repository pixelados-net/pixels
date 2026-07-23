// Package invite encodes GAMEINVITE responses.
package invite

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAMEINVITE.
const Header uint16 = 904

// Encode creates one GAMEINVITE response.
func Encode(gameTypeID int32, inviterID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(gameTypeID), codec.Int32(inviterID))
}
