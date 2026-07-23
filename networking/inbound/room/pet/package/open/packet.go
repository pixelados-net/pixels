// Package open decodes PET_OPEN_PACKAGE requests.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_OPEN_PACKAGE.
const Header uint16 = 3698

// Definition describes PET_OPEN_PACKAGE fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField}

// Payload contains one package opening request.
type Payload struct {
	// ObjectID identifies the package furniture item.
	ObjectID int64
	// Name stores the requested pet name.
	Name string
}

// Decode decodes PET_OPEN_PACKAGE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ObjectID: int64(values[0].Int32), Name: values[1].String}, nil
}
