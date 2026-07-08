// Package flooritems contains the ROOM_FLOOR_ITEMS outbound packet.
package flooritems

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_FLOOR_ITEMS packet identifier.
	Header uint16 = 1778

	// defaultKind is the gift/song/default value for non-gift, non-music-disc items.
	defaultKind int32 = 1

	// nonLimitedFlag is the extradata flag value for non-limited items.
	nonLimitedFlag int32 = 0

	// unknownExpiration is the constant expiration/unknown trailing value.
	unknownExpiration int32 = -1
)

// Owner stores one floor item owner name record.
type Owner struct {
	// ID stores the durable owner player id.
	ID int64

	// Name stores the visible owner name.
	Name string
}

// FloorItem stores one placed floor item record.
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

	// UsagePolicy stores the item interaction usage policy.
	UsagePolicy int32

	// OwnerID stores the durable owner player id.
	OwnerID int64
}

// Encode creates a ROOM_FLOOR_ITEMS packet.
func Encode(owners []Owner, items []FloorItem) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(owners))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, owner := range owners {
		payload, err = appendOwner(payload, owner)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(items))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, item := range items {
		payload, err = appendFloorItem(payload, item)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendOwner appends one owner name record.
func appendOwner(dst []byte, owner Owner) ([]byte, error) {
	return codec.AppendPayload(dst, ownerDefinition(),
		codec.Int32(int32(owner.ID)),
		codec.String(owner.Name),
	)
}

// appendFloorItem appends one floor item record.
func appendFloorItem(dst []byte, item FloorItem) ([]byte, error) {
	return codec.AppendPayload(dst, floorItemDefinition(),
		codec.Int32(int32(item.ID)),
		codec.Int32(int32(item.SpriteID)),
		codec.Int32(int32(item.X)),
		codec.Int32(int32(item.Y)),
		codec.Int32(int32(item.Rotation)),
		codec.String(item.Z),
		codec.String(item.ExtraHeight),
		codec.Int32(defaultKind),
		codec.Int32(nonLimitedFlag),
		codec.String(item.ExtraData),
		codec.Int32(unknownExpiration),
		codec.Int32(item.UsagePolicy),
		codec.Int32(int32(item.OwnerID)),
	)
}

// ownerDefinition returns the owner name field order.
func ownerDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("ownerId", codec.Int32Field),
		codec.Named("ownerName", codec.StringField),
	}
}

// floorItemDefinition returns the floor item field order.
func floorItemDefinition() codec.Definition {
	return codec.Definition{
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
}
