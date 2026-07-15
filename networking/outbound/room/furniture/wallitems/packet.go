// Package wallitems encodes the complete room wall-item snapshot.
package wallitems

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the ITEM_WALL identifier.
const Header uint16 = 1369

// Owner stores one wall-item owner name.
type Owner struct {
	// ID identifies the owner.
	ID int64
	// Name stores the visible owner name.
	Name string
}

// Item stores one placed wall item.
type Item struct {
	// ID identifies the durable furniture item.
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
}

// Encode creates an ITEM_WALL packet.
func Encode(owners []Owner, items []Item) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(owners))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, owner := range owners {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(owner.ID)), codec.String(owner.Name))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(items))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, item := range items {
		payload, err = appendItem(payload, item)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendItem appends Nitro's compact wall-item record.
func appendItem(payload []byte, item Item) ([]byte, error) {
	return codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.String(intString(item.ID)), codec.Int32(int32(item.SpriteID)), codec.String(item.WallPosition), codec.String(item.ExtraData), codec.Int32(-1), codec.Int32(item.UsagePolicy), codec.Int32(int32(item.OwnerID)))
}

// intString formats a durable id without allocation-heavy reflection.
func intString(value int64) string { return strconv.FormatInt(value, 10) }
