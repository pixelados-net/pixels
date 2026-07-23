// Package rolling contains the ROOM_ROLLING outbound packet.
package rolling

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_ROLLING packet identifier.
	Header uint16 = 3207
	// movementSlide asks Nitro to settle the rolled unit in standing posture.
	movementSlide int32 = 2
)

// Unit stores one unit animated by a roller.
type Unit struct {
	// RoomIndex stores the room-local unit id.
	RoomIndex int64
	// FromZ stores the source unit height.
	FromZ string
	// ToZ stores the destination unit height.
	ToZ string
}

// Item stores one floor item animated by a roller.
type Item struct {
	// ID stores the durable furniture item id.
	ID int64
	// FromZ stores the source item height.
	FromZ string
	// ToZ stores the destination item height.
	ToZ string
}

// Option configures optional ROOM_ROLLING fields.
type Option func(*options)

// options stores optional packet data.
type options struct {
	// unit stores the optional single unit supported by Nitro's parser.
	unit *Unit
}

// WithUnit appends one rolled unit after the roller id.
func WithUnit(unit Unit) Option {
	return func(options *options) {
		options.unit = &unit
	}
}

// Encode creates the legacy ROOM_ROLLING shape consumed by the bundled Nitro renderer.
func Encode(fromX int, fromY int, targetX int, targetY int, items []Item, rollerID int64, packetOptions ...Option) (codec.Packet, error) {
	configured := options{}
	for _, option := range packetOptions {
		option(&configured)
	}
	definition, values := payloadShape(len(items), configured.unit != nil)
	values[0], values[1] = codec.Int32(int32(fromX)), codec.Int32(int32(fromY))
	values[2], values[3] = codec.Int32(int32(targetX)), codec.Int32(int32(targetY))
	values[4] = codec.Int32(int32(len(items)))
	offset := 5
	for index, item := range items {
		base := offset + index*3
		values[base] = codec.Int32(int32(item.ID))
		values[base+1], values[base+2] = codec.String(item.FromZ), codec.String(item.ToZ)
	}
	offset += len(items) * 3
	values[offset] = codec.Int32(int32(rollerID))
	if configured.unit != nil {
		values[offset+1] = codec.Int32(movementSlide)
		values[offset+2] = codec.Int32(int32(configured.unit.RoomIndex))
		values[offset+3] = codec.String(configured.unit.FromZ)
		values[offset+4] = codec.String(configured.unit.ToZ)
	}
	return codec.NewPacket(Header, definition, values...)
}

// payloadShape builds the legacy dynamic item list and optional unit tail.
func payloadShape(itemCount int, hasUnit bool) (codec.Definition, []codec.Value) {
	count := 6 + itemCount*3
	if hasUnit {
		count += 4
	}
	definition := make(codec.Definition, count)
	for index := range 5 {
		definition[index] = codec.Int32Field
	}
	offset := 5
	for range itemCount {
		definition[offset], definition[offset+1], definition[offset+2] = codec.Int32Field, codec.StringField, codec.StringField
		offset += 3
	}
	definition[offset] = codec.Int32Field
	if hasUnit {
		definition[offset+1], definition[offset+2] = codec.Int32Field, codec.Int32Field
		definition[offset+3], definition[offset+4] = codec.StringField, codec.StringField
	}
	return definition, make([]codec.Value, count)
}
