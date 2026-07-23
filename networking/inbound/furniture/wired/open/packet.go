// Package open decodes the WIRED_OPEN inbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_OPEN packet identifier.
const Header uint16 = 768

// Payload stores one WIRED editor request.
type Payload struct {
	// ItemID identifies the WIRED furniture item.
	ItemID int32
}

// Definition describes WIRED_OPEN.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode decodes one WIRED_OPEN packet.
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
