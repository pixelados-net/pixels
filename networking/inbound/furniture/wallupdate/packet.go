// Package wallupdate decodes wall furniture repositioning.
package wallupdate

import "github.com/niflaot/pixels/networking/codec"

// Header is the FURNITURE_WALL_UPDATE identifier.
const Header uint16 = 168

// Payload contains one wall placement update.
type Payload struct {
	// ItemID identifies the placed wall furniture.
	ItemID int32
	// WallPosition stores Nitro's modern wall coordinates.
	WallPosition string
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField}

// Decode decodes one wall furniture update.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, WallPosition: values[1].String}, nil
}
