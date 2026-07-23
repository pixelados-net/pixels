// Package loadurl encodes LOAD_GAME_URL responses.
package loadurl

import "github.com/niflaot/pixels/networking/codec"

// Header identifies LOAD_GAME_URL.
const Header uint16 = 2624

// Encode creates one LOAD_GAME_URL response.
func Encode(gameTypeID int32, gameClientID string, url string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(gameTypeID), codec.String(gameClientID), codec.String(url))
}
