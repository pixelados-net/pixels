// Package name contains the UNIT_CHANGE_NAME outbound packet.
package name

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_CHANGE_NAME.
const Header uint16 = 2182

// Definition describes UNIT_CHANGE_NAME fields.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("roomUnitId", codec.Int32Field), codec.Named("username", codec.StringField)}

// Encode creates a UNIT_CHANGE_NAME packet.
func Encode(playerID int32, roomUnitID int32, username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(playerID), codec.Int32(roomUnitID), codec.String(username))
}
