// Package update contains the FLOOR_ITEM_UPDATE outbound packet.
package update

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

const (
	// Header is the FLOOR_ITEM_UPDATE packet identifier.
	Header uint16 = 3776

	// updateKind is the constant gift/song/default value used on update, matching the real protocol.
	updateKind int32 = 0

	// nonLimitedFlag is the extradata flag value for non-limited items.
	nonLimitedFlag int32 = 0

	// unknownExpiration is the constant expiration/unknown trailing value.
	unknownExpiration int32 = -1

	// updateUsagePolicy is the constant usage policy value used on update, matching the real protocol.
	updateUsagePolicy int32 = 0
)

// FloorItem stores one moved, rotated, or state-changed floor item record.
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

	// OwnerID stores the durable owner player id.
	OwnerID int64
}

// Definition describes the FLOOR_ITEM_UPDATE payload fields.
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
}

// Encode creates a FLOOR_ITEM_UPDATE packet.
func Encode(item FloorItem) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, prefixDefinition(),
		codec.Int32(int32(item.ID)),
		codec.Int32(int32(item.SpriteID)),
		codec.Int32(int32(item.X)),
		codec.Int32(int32(item.Y)),
		codec.Int32(int32(item.Rotation)),
		codec.String(item.Z),
		codec.String(item.ExtraHeight),
		codec.Int32(updateKind),
	)
	if err != nil {
		return codec.Packet{}, err
	}
	if item.Data != nil {
		payload, err = item.Data.Append(payload)
	} else {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(nonLimitedFlag), codec.String(item.ExtraData))
	}
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, suffixDefinition(),
		codec.Int32(unknownExpiration),
		codec.Int32(updateUsagePolicy),
		codec.Int32(int32(item.OwnerID)))
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// prefixDefinition returns the fields before object data.
func prefixDefinition() codec.Definition {
	return Definition[:8]
}

// suffixDefinition returns the fields after object data.
func suffixDefinition() codec.Definition {
	return Definition[10:]
}
