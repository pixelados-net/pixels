// Package favoriteupdate contains one Nitro social-group outbound packet.
package favoriteupdate

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3403

// Encode creates the complete packet.
func Encode(roomIndex int32, groupID int64, status int32, groupName string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}, codec.Int32(int32(roomIndex)), codec.Int32(int32(groupID)), codec.Int32(int32(status)), codec.String(groupName))
}
