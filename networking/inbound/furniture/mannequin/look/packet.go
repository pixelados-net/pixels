// Package look decodes mannequin look saves.
package look

import "github.com/niflaot/pixels/networking/codec"

// Header is the MANNEQUIN_SAVE_LOOK identifier.
const Header uint16 = 2209

// Payload identifies one mannequin.
type Payload struct { // ItemID identifies the mannequin.
	ItemID int32
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a mannequin look save.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32}, nil
}
