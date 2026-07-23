// Package like contains the ROOM_LIKE inbound packet.
package like

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_LIKE.
	Header uint16 = 3582
)

// Payload contains unpacked room like fields.
type Payload struct {
	// Rating stores Nitro's positive rating value.
	Rating int32
}

// Definition describes ROOM_LIKE fields.
var Definition = codec.Definition{codec.Named("rating", codec.Int32Field)}

// Decode unpacks a ROOM_LIKE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Rating: values[0].Int32}, nil
}
