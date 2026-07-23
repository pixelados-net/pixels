// Package remove contains the REMOVE_FLOOR_ITEM outbound packet.
package remove

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the REMOVE_FLOOR_ITEM packet identifier.
	Header uint16 = 2703
)

// Definition describes the REMOVE_FLOOR_ITEM payload fields.
var Definition = codec.Definition{
	codec.Named("id", codec.StringField),
	codec.Named("expired", codec.BooleanField),
	codec.Named("ownerId", codec.Int32Field),
	codec.Named("unknown", codec.Int32Field),
}

// Encode creates a REMOVE_FLOOR_ITEM packet.
func Encode(itemID int64, ownerID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.String(strconv.FormatInt(itemID, 10)),
		codec.Bool(false),
		codec.Int32(int32(ownerID)),
		codec.Int32(0),
	)
}
