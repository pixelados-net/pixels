// Package add contains the ADD_FLOOR_ITEM outbound packet.
package add

import "github.com/niflaot/pixels/networking/codec"

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

	// UsagePolicy stores the item interaction usage policy.
	UsagePolicy int32

	// OwnerID stores the durable owner player id.
	OwnerID int64

	// OwnerName stores the visible owner name.
	OwnerName string
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
	return codec.NewPacket(Header, Definition,
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
		codec.String(item.OwnerName),
	)
}
