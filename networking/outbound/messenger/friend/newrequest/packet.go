// Package newrequest contains the MESSENGER_REQUEST outbound packet.
package newrequest

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_REQUEST.
const Header uint16 = 2219

// Encode creates one live incoming friend-request notification.
func Encode(playerID int64, username string, look string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(int32(playerID)), codec.String(username), codec.String(look))
}
