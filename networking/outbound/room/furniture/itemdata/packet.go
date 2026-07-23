// Package itemdata encodes editable wall-item content.
package itemdata

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the FURNITURE_ITEMDATA identifier.
const Header uint16 = 2202

// Encode creates a FURNITURE_ITEMDATA packet.
func Encode(itemID int64, data string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.StringField}, codec.String(strconv.FormatInt(itemID, 10)), codec.String(data))
}
