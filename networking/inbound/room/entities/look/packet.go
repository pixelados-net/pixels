// Package look contains the UNIT_LOOK inbound packet.
package look

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the UNIT_LOOK packet identifier.
	Header uint16 = 3301
)

// Payload contains the unpacked UNIT_LOOK fields.
type Payload struct {
	// X stores the target look-at x coordinate.
	X int32

	// Y stores the target look-at y coordinate.
	Y int32
}

// Definition describes the UNIT_LOOK payload fields.
var Definition = codec.Definition{
	codec.Named("x", codec.Int32Field),
	codec.Named("y", codec.Int32Field),
}

// Decode unpacks a UNIT_LOOK packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{X: values[0].Int32, Y: values[1].Int32}, nil
}
