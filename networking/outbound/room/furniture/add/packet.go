// Package add contains the ADD_FLOOR_ITEM outbound packet.
package add

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

const (
	// Header is the ADD_FLOOR_ITEM packet identifier.
	Header uint16 = 1534

	// defaultKind is the gift/song/default value for non-gift, non-music-disc items.
	defaultKind int32 = 1

	// nonLimitedFlag is the extradata flag value for non-limited items.
	nonLimitedFlag int32 = 0

	// unknownExpiration is the constant expiration/unknown trailing value.
	unknownExpiration int32 = -1
)

// FloorItem stores one newly placed floor item record.
type FloorItem struct {
	// ID stores the durable furniture item id.
	ID int64

	// SpriteID stores the Nitro rendering class id.
	SpriteID int

	// X stores the floor tile x coordinate.
	X int

	// Y stores the floor tile y coordinate.
	Y int

	// Rotation stores the floor instance rotation.
	Rotation int

	// Z stores the resolved placement height.
	Z string

	// ExtraHeight stores the walkable top height, or an empty string when not applicable.
	ExtraHeight string

	// ExtraData stores simple protocol-facing visual state.
	ExtraData string

	// Data stores specialized furniture object data.
	Data *stuffdata.Data

	// UsagePolicy stores the item interaction usage policy.
	UsagePolicy int32

	// Kind stores the packed gift variant or the regular default value.
	Kind int32

	// GiftWrapped reports whether Kind may intentionally be zero.
	GiftWrapped bool

	// OwnerID stores the durable owner player id.
	OwnerID int64

	// OwnerName stores the visible owner name.
	OwnerName string

	// GiftMessage stores the gift tag text for wrapped presents.
	GiftMessage string

	// GiftProductCode stores the wrapped product code or furniture class name.
	GiftProductCode string

	// GiftSenderName stores the visible gift sender name.
	GiftSenderName string

	// GiftSenderFigure stores the visible gift sender figure.
	GiftSenderFigure string
}

// Definition describes the ADD_FLOOR_ITEM payload fields.
var Definition = codec.Definition{
	codec.Named("id", codec.Int32Field),
	codec.Named("spriteId", codec.Int32Field),
	codec.Named("x", codec.Int32Field),
	codec.Named("y", codec.Int32Field),
	codec.Named("rotation", codec.Int32Field),
	codec.Named("z", codec.StringField),
	codec.Named("extraHeight", codec.StringField),
	codec.Named("kind", codec.Int32Field),
	codec.Named("limitedFlag", codec.Int32Field),
	codec.Named("extradata", codec.StringField),
	codec.Named("expiration", codec.Int32Field),
	codec.Named("usagePolicy", codec.Int32Field),
	codec.Named("ownerId", codec.Int32Field),
	codec.Named("ownerName", codec.StringField),
}

// Encode creates an ADD_FLOOR_ITEM packet.
func Encode(item FloorItem) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, prefixDefinition(),
		codec.Int32(int32(item.ID)),
		codec.Int32(int32(item.SpriteID)),
		codec.Int32(int32(item.X)),
		codec.Int32(int32(item.Y)),
		codec.Int32(int32(item.Rotation)),
		codec.String(item.Z),
		codec.String(item.ExtraHeight),
		codec.Int32(itemKind(item)),
	)
	if err != nil {
		return codec.Packet{}, err
	}
	if item.GiftWrapped {
		payload, err = appendGiftData(payload, item)
	} else if item.Data != nil {
		payload, err = item.Data.Append(payload)
	} else {
		payload, err = codec.AppendPayload(payload, regularDataDefinition(),
			codec.Int32(nonLimitedFlag), codec.String(item.ExtraData))
	}
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, suffixDefinition(),
		codec.Int32(unknownExpiration),
		codec.Int32(item.UsagePolicy),
		codec.Int32(int32(item.OwnerID)),
		codec.String(item.OwnerName),
	)
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// itemKind resolves the zero value to the normal furniture kind.
func itemKind(item FloorItem) int32 {
	if item.GiftWrapped || item.Kind != 0 {
		return item.Kind
	}

	return defaultKind
}

// appendGiftData appends Nitro's map object data for present tags.
func appendGiftData(dst []byte, item FloorItem) ([]byte, error) {
	return stuffdata.AppendMap(dst, []stuffdata.Pair{
		{Key: "EXTRA_PARAM", Value: ""},
		{Key: "MESSAGE", Value: item.GiftMessage},
		{Key: "PURCHASER_NAME", Value: item.GiftSenderName},
		{Key: "PURCHASER_FIGURE", Value: item.GiftSenderFigure},
		{Key: "PRODUCT_CODE", Value: item.GiftProductCode},
		{Key: "state", Value: giftState(item.ExtraData)},
	})
}

// giftState resolves the present visual state value.
func giftState(value string) string {
	if value == "" {
		return "0"
	}

	return value
}

// prefixDefinition returns the floor item fields before object data.
func prefixDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("id", codec.Int32Field),
		codec.Named("spriteId", codec.Int32Field),
		codec.Named("x", codec.Int32Field),
		codec.Named("y", codec.Int32Field),
		codec.Named("rotation", codec.Int32Field),
		codec.Named("z", codec.StringField),
		codec.Named("extraHeight", codec.StringField),
		codec.Named("kind", codec.Int32Field),
	}
}

// regularDataDefinition returns the legacy simple object-data fields.
func regularDataDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("limitedFlag", codec.Int32Field),
		codec.Named("extradata", codec.StringField),
	}
}

// suffixDefinition returns the floor item fields after object data.
func suffixDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("expiration", codec.Int32Field),
		codec.Named("usagePolicy", codec.Int32Field),
		codec.Named("ownerId", codec.Int32Field),
		codec.Named("ownerName", codec.StringField),
	}
}
