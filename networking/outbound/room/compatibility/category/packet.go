// Package category encodes the unused SHOW_ENFORCE_ROOM_CATEGORY packet.
package category

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SHOW_ENFORCE_ROOM_CATEGORY.
const Header uint16 = 3896

// Definition describes the unused selection type.
var Definition = codec.Definition{codec.Named("selectionType", codec.Int32Field)}

// Encode creates a compatibility category dialog packet.
func Encode(selectionType int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(selectionType))
}
