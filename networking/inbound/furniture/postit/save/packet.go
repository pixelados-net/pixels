// Package save decodes initial sticky-pole post-it content.
package save

import "github.com/niflaot/pixels/networking/codec"

// Header is the FURNITURE_POSTIT_SAVE_STICKY_POLE identifier.
const Header uint16 = 3283

// Payload contains initial post-it fields.
type Payload struct {
	// ItemID identifies the newly placed post-it.
	ItemID int32
	// WallPosition stores Nitro wall coordinates.
	WallPosition string
	// Color stores the post-it hexadecimal color.
	Color string
	// Text stores user-authored content.
	Text string
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField}

// Decode decodes initial post-it content.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, WallPosition: values[1].String, Color: values[2].String, Text: values[3].String}, nil
}
