// Package thumbnailupdate contains UPDATE_ROOM_THUMBNAIL.
package thumbnailupdate

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UPDATE_ROOM_THUMBNAIL.
const Header uint16 = 2468

// Definition describes the compatibility payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Payload stores the legacy thumbnail composition request.
type Payload struct {
	// FlatID identifies the room.
	FlatID int32
	// BackgroundImageID identifies the background asset.
	BackgroundImageID int32
	// ForegroundImageID identifies the foreground asset.
	ForegroundImageID int32
	// ObjectCount stores the legacy object count.
	ObjectCount int32
}

// Decode decodes the compatibility request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{FlatID: values[0].Int32, BackgroundImageID: values[1].Int32, ForegroundImageID: values[2].Int32, ObjectCount: values[3].Int32}, nil
}
