// Package name decodes mannequin outfit-name saves.
package name

import "github.com/niflaot/pixels/networking/codec"

// Header is the MANNEQUIN_SAVE_NAME identifier.
const Header uint16 = 2850

// Payload contains mannequin naming fields.
type Payload struct {
	// ItemID identifies the mannequin.
	ItemID int32
	// Name stores the visible outfit label.
	Name string
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField}

// Decode decodes a mannequin name save.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Name: values[1].String}, nil
}
