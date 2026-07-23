// Package wallmultistate decodes the FURNITURE_WALL_MULTISTATE inbound packet.
package wallmultistate

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FURNITURE_WALL_MULTISTATE.
const Header uint16 = 210

// Definition describes the wall furniture interaction fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
	codec.Named("state", codec.Int32Field),
}

// Payload stores one wall furniture interaction.
type Payload struct {
	// ItemID identifies the clicked wall furniture item.
	ItemID int32
	// State stores the client interaction state parameter.
	State int32
}

// Decode decodes one FURNITURE_WALL_MULTISTATE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, State: values[1].Int32}, nil
}
