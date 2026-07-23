// Package walladd encodes a newly placed wall item.
package walladd

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the ITEM_WALL_ADD identifier.
const Header uint16 = 2187

// Item stores one newly placed wall item.
type Item struct {
	// ID identifies the durable item.
	ID int64
	// SpriteID identifies its Nitro furniture class.
	SpriteID int
	// WallPosition stores modern wall coordinates.
	WallPosition string
	// ExtraData stores wall-item state.
	ExtraData string
	// UsagePolicy stores whether the item is usable.
	UsagePolicy int32
	// OwnerID identifies the owner.
	OwnerID int64
	// OwnerName stores the visible owner name.
	OwnerName string
}

// Encode creates an ITEM_WALL_ADD packet.
func Encode(item Item) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}, codec.String(strconv.FormatInt(item.ID, 10)), codec.Int32(int32(item.SpriteID)), codec.String(item.WallPosition), codec.String(item.ExtraData), codec.Int32(-1), codec.Int32(item.UsagePolicy), codec.Int32(int32(item.OwnerID)), codec.String(item.OwnerName))
}
