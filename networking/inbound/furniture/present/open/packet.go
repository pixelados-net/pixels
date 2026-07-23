// Package open contains the OPEN_PRESENT inbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the OPEN_PRESENT packet identifier.
	Header uint16 = 3558
)

// Payload stores one present open request.
type Payload struct {
	// ItemID identifies the placed gift furniture item.
	ItemID int32
}

// Definition describes the OPEN_PRESENT payload fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
}

// Decode decodes one OPEN_PRESENT packet.
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
