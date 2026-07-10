// Package list contains the FURNITURE_INVENTORY outbound packet.
package list

import (
	"errors"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the FURNITURE_INVENTORY packet identifier.
	Header uint16 = 994

	// floorTypeCode is the protocol type code for floor furniture.
	floorTypeCode = "S"

	// wallTypeCode is the protocol type code for wall furniture.
	wallTypeCode = "I"

	// nonLimitedFlag is the extradata flag value for non-limited items.
	nonLimitedFlag int32 = 0

	// unknownTrailer is a constant trailing field observed in the real protocol.
	unknownTrailer int32 = -1

	// floorTrailerKind is the floor-specific trailing gift/default value.
	floorTrailerKind int32 = 1
)

var (
	// ErrUnsupportedKind reports an inventory kind unsupported by this packet.
	ErrUnsupportedKind = errors.New("unsupported inventory furniture kind")
)

// Kind identifies how Nitro represents an inventory furniture item.
type Kind string

const (
	// KindFloor identifies a floor furniture item and is also the zero-value behavior.
	KindFloor Kind = "floor"

	// KindWall identifies a wall furniture item.
	KindWall Kind = "wall"
)

// Category identifies Nitro's special inventory furniture treatment.
type Category int32

const (
	// CategoryDefault identifies regular furniture and is the zero-value behavior.
	CategoryDefault Category = 1

	// CategoryWallpaper identifies wallpaper consumables.
	CategoryWallpaper Category = 2

	// CategoryFloor identifies floor paint consumables.
	CategoryFloor Category = 3

	// CategoryLandscape identifies landscape consumables.
	CategoryLandscape Category = 4
)

// Item stores one inventory furniture item record.
type Item struct {
	// ID stores the durable furniture item id.
	ID int64

	// SpriteID stores the Nitro rendering class id.
	SpriteID int

	// Kind identifies whether the item belongs on the floor or wall.
	Kind Kind

	// Category identifies special inventory treatment such as room paint.
	Category Category

	// ExtraData stores simple protocol-facing visual state.
	ExtraData string

	// AllowInventoryStack reports whether inventory can group identical items.
	AllowInventoryStack bool
}

// Encode creates a FURNITURE_INVENTORY packet for one fragment of a player's inventory.
func Encode(fragmentNumber int, totalFragments int, items []Item) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, headerDefinition(),
		codec.Int32(int32(totalFragments)),
		codec.Int32(int32(fragmentNumber-1)),
		codec.Int32(int32(len(items))),
	)
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

// appendItem appends one inventory furniture item record.
func appendItem(dst []byte, item Item) ([]byte, error) {
	typeCode, err := itemTypeCode(item.Kind)
	if err != nil {
		return nil, err
	}
	dst, err = codec.AppendPayload(dst, itemDefinition(),
		codec.Int32(int32(item.ID)),
		codec.String(typeCode),
		codec.Int32(int32(item.ID)),
		codec.Int32(int32(item.SpriteID)),
		codec.Int32(int32(itemCategory(item.Category))),
		codec.Int32(nonLimitedFlag),
		codec.String(item.ExtraData),
		codec.Bool(false),
		codec.Bool(false),
		codec.Bool(item.AllowInventoryStack),
		codec.Bool(false),
		codec.Int32(unknownTrailer),
		codec.Bool(true),
		codec.Int32(unknownTrailer),
	)
	if err != nil || typeCode == wallTypeCode {
		return dst, err
	}

	return codec.AppendPayload(dst, floorDefinition(), codec.String(""), codec.Int32(floorTrailerKind))
}

// itemCategory resolves the zero value to Nitro's regular furniture category.
func itemCategory(category Category) Category {
	if category == 0 {
		return CategoryDefault
	}

	return category
}

// itemTypeCode maps an inventory kind to Nitro's wire discriminator.
func itemTypeCode(kind Kind) (string, error) {
	switch kind {
	case "", KindFloor:
		return floorTypeCode, nil
	case KindWall:
		return wallTypeCode, nil
	default:
		return "", ErrUnsupportedKind
	}
}

// headerDefinition returns the fragment header field order.
func headerDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("totalFragments", codec.Int32Field),
		codec.Named("fragmentNumber", codec.Int32Field),
		codec.Named("itemCount", codec.Int32Field),
	}
}

// itemDefinition returns the inventory item field order.
func itemDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("giftAdjustedId", codec.Int32Field),
		codec.Named("typeCode", codec.StringField),
		codec.Named("id", codec.Int32Field),
		codec.Named("spriteId", codec.Int32Field),
		codec.Named("kind", codec.Int32Field),
		codec.Named("limitedFlag", codec.Int32Field),
		codec.Named("extradata", codec.StringField),
		codec.Named("allowRecycle", codec.BooleanField),
		codec.Named("allowTrade", codec.BooleanField),
		codec.Named("allowInventoryStack", codec.BooleanField),
		codec.Named("allowMarketplace", codec.BooleanField),
		codec.Named("unknown1", codec.Int32Field),
		codec.Named("hasRentPeriodStarted", codec.BooleanField),
		codec.Named("unknown2", codec.Int32Field),
	}
}

// floorDefinition returns fields present only on floor inventory items.
func floorDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("songId", codec.StringField),
		codec.Named("floorTrailerKind", codec.Int32Field),
	}
}
