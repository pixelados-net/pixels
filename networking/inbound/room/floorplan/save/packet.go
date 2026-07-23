// Package save contains the ROOM_MODEL_SAVE inbound packet.
package save

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_MODEL_SAVE.
	Header uint16 = 875
)

// Payload contains unpacked custom floor plan fields.
type Payload struct {
	// Heightmap stores compact room geometry.
	Heightmap string
	// DoorX stores the entry tile x coordinate.
	DoorX int32
	// DoorY stores the entry tile y coordinate.
	DoorY int32
	// DoorDirection stores entry rotation.
	DoorDirection int32
	// WallThickness stores wall rendering thickness.
	WallThickness int32
	// FloorThickness stores floor rendering thickness.
	FloorThickness int32
	// WallHeight stores fixed wall height or -1 for automatic height.
	WallHeight int32
}

// Definition describes ROOM_MODEL_SAVE fields in Nitro wire order.
var Definition = codec.Definition{
	codec.Named("heightmap", codec.StringField),
	codec.Named("doorX", codec.Int32Field),
	codec.Named("doorY", codec.Int32Field),
	codec.Named("doorDirection", codec.Int32Field),
	codec.Named("wallThickness", codec.Int32Field),
	codec.Named("floorThickness", codec.Int32Field),
	codec.Named("wallHeight", codec.Int32Field),
}

// Decode unpacks a ROOM_MODEL_SAVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}

	return Payload{
		Heightmap: values[0].String, DoorX: values[1].Int32, DoorY: values[2].Int32,
		DoorDirection: values[3].Int32, WallThickness: values[4].Int32,
		FloorThickness: values[5].Int32, WallHeight: values[6].Int32,
	}, nil
}
