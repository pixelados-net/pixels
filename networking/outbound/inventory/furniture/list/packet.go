// Package list contains the FURNITURE_INVENTORY outbound packet.
package list

import (
	"errors"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
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

	// Data stores optional specialized protocol-facing object data.
	Data *stuffdata.Data

	// AllowInventoryStack reports whether inventory can group identical items.
	AllowInventoryStack bool

	// GiftWrapped reports whether Nitro should expose gift wrapping variants.
	GiftWrapped bool

	// AllowTrade reports whether Nitro may offer the item in direct trades.
	AllowTrade bool

	// AllowMarketplace reports whether Nitro may list the item for sale.
	AllowMarketplace bool

	// AllowRecycle reports whether Nitro may submit the item to the recycler.
	AllowRecycle bool

	// LimitedEditionNumber stores the optional LTD serial number.
	LimitedEditionNumber *int32

	// GiftBoxID stores the selected gift box variant.
	GiftBoxID int32

	// GiftRibbonID stores the selected gift ribbon variant.
	GiftRibbonID int32
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
		payload, err = AppendItem(payload, item)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// AppendItem appends one inventory furniture item record.
func AppendItem(dst []byte, item Item) ([]byte, error) {
	typeCode, err := itemTypeCode(item.Kind)
	if err != nil {
		return nil, err
	}
	adjustedID := int32(item.ID)
	kind := itemKind(item)
	dst, err = codec.AppendPayload(dst, itemPrefixDefinition(),
		codec.Int32(adjustedID),
		codec.String(typeCode),
		codec.Int32(int32(item.ID)),
		codec.Int32(int32(item.SpriteID)),
		codec.Int32(kind),
	)
	if err != nil {
		return nil, err
	}
	if item.LimitedEditionNumber != nil {
		dst, err = codec.AppendPayload(dst, limitedDataDefinition(), codec.Int32(1), codec.Int32(*item.LimitedEditionNumber), codec.Int32(0))
	} else if item.Data != nil {
		dst, err = item.Data.Append(dst)
	} else {
		dst, err = codec.AppendPayload(dst, regularDataDefinition(), codec.Int32(nonLimitedFlag), codec.String(item.ExtraData))
	}
	if err != nil {
		return nil, err
	}
	dst, err = codec.AppendPayload(dst, itemSuffixDefinition(),
		codec.Bool(item.AllowRecycle),
		codec.Bool(item.AllowTrade),
		codec.Bool(item.AllowInventoryStack),
		codec.Bool(item.AllowMarketplace),
		codec.Int32(unknownTrailer),
		codec.Bool(true),
		codec.Int32(unknownTrailer),
	)
	if err != nil || typeCode == wallTypeCode {
		return dst, err
	}

	return codec.AppendPayload(dst, floorDefinition(), codec.String(""), codec.Int32(kind))
}

// itemKind resolves the normal category or packed gift box and ribbon variant.
func itemKind(item Item) int32 {
	if item.GiftWrapped {
		return item.GiftBoxID*1000 + item.GiftRibbonID
	}

	return int32(itemCategory(item.Category))
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
