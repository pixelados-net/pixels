// Package use decodes the FURNITURE_MULTISTATE inbound packet.
package use

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FURNITURE_MULTISTATE packet identifier.
	Header uint16 = 99
)

// Payload stores one floor furniture use request.
type Payload struct {
	// ItemID identifies the clicked furniture item.
	ItemID int32
	// State stores the client interaction state parameter.
	State int32
}

// Definition describes the FURNITURE_MULTISTATE payload fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
	codec.Named("state", codec.Int32Field),
}

// Decode decodes one FURNITURE_MULTISTATE packet.
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
