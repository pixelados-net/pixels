// Package trophy decodes MYSTERYBOX_OPEN_TROPHY requests.
package trophy

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MYSTERYBOX_OPEN_TROPHY.
const Header uint16 = 3074

// Payload contains one trophy inscription.
type Payload struct {
	// ItemID identifies the mystery trophy.
	ItemID int32
	// Text stores the requested inscription.
	Text string
}

// Definition describes the trophy inscription request.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("text", codec.StringField)}

// Decode returns one trophy inscription.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Text: values[1].String}, nil
}
