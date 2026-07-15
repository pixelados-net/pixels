// Package wallupdate encodes a wall-item state refresh.
package wallupdate

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the ITEM_WALL_UPDATE identifier.
const Header uint16 = 2009

// Encode creates an ITEM_WALL_UPDATE packet.
func Encode(itemID int64, spriteID int, wallPosition string, extraData string, usagePolicy int32, ownerID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.String(strconv.FormatInt(itemID, 10)), codec.Int32(int32(spriteID)), codec.String(wallPosition), codec.String(extraData), codec.Int32(-1), codec.Int32(usagePolicy), codec.Int32(int32(ownerID)))
}
