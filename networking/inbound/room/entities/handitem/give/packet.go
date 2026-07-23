// Package give decodes giving a carried hand item.
package give

import "github.com/niflaot/pixels/networking/codec"

// Header is the UNIT_GIVE_HANDITEM identifier.
const Header uint16 = 2941

// Payload identifies the target room-local unit.
type Payload struct { // UnitID identifies the recipient unit.
	UnitID int32
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes giving a carried hand item.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{UnitID: values[0].Int32}, nil
}
