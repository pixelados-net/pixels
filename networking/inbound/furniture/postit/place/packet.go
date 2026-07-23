// Package place decodes post-it wall placement.
package place

import "github.com/niflaot/pixels/networking/codec"

// Header is the FURNITURE_POSTIT_PLACE identifier.
const Header uint16 = 2248

// Payload contains post-it placement fields.
type Payload struct {
	// ItemID identifies the inventory post-it.
	ItemID int32
	// WallPosition stores Nitro wall coordinates.
	WallPosition string
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField}

// Decode decodes a post-it placement.
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
