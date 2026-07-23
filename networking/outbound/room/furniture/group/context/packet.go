// Package contextmenu contains one Nitro social-group outbound packet.
package contextmenu

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3293

// Encode creates the complete packet.
func Encode(objectID int64, groupID int64, groupName string, homeRoomID int64, member bool, forumReadable bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField, codec.BooleanField}, codec.Int32(int32(objectID)), codec.Int32(int32(groupID)), codec.String(groupName), codec.Int32(int32(homeRoomID)), codec.Bool(member), codec.Bool(forumReadable))
}
