// Package wallremove encodes removal of one wall item.
package wallremove

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the ITEM_WALL_REMOVE identifier.
const Header uint16 = 3208

// Encode creates an ITEM_WALL_REMOVE packet.
func Encode(itemID int64, ownerID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(strconv.FormatInt(itemID, 10)), codec.Int32(int32(ownerID)))
}
