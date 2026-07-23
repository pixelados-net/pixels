// Package unload encodes UNLOADGAME responses.
package unload

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNLOADGAME.
const Header uint16 = 1715

// Encode creates one UNLOADGAME response.
func Encode(gameTypeID int32, gameClientID string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(gameTypeID), codec.String(gameClientID))
}
