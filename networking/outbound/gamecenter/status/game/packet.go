// Package game encodes GAMESTATUSMESSAGE responses.
package game

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAMESTATUSMESSAGE.
const Header uint16 = 3805

// Encode creates one GAMESTATUSMESSAGE response.
func Encode(gameTypeID int32, status int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(gameTypeID), codec.Int32(status))
}
