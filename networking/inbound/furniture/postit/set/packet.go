// Package set decodes a post-it data update.
package set

import "github.com/niflaot/pixels/networking/codec"

// Header is the SET_ITEM_DATA identifier.
const Header uint16 = 3666

// Payload contains editable post-it fields.
type Payload struct {
	// ItemID identifies the wall item.
	ItemID int32
	// Color stores a hexadecimal post-it color.
	Color string
	// Text stores user-authored content.
	Text string
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}

// Decode decodes a post-it data update.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Color: values[1].String, Text: values[2].String}, nil
}
