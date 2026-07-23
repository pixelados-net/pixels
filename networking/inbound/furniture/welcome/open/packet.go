// Package open contains the retired WELCOME_OPEN_GIFT inbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WELCOME_OPEN_GIFT.
const Header uint16 = 2638

// Definition describes the WELCOME_OPEN_GIFT payload.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Payload contains decoded WELCOME_OPEN_GIFT fields.
type Payload struct {
	// ItemID identifies the ignored legacy welcome furniture.
	ItemID int32
}

// Decode decodes a retired welcome gift open request.
//
// Deprecated: the legacy welcome-gift journey is intentionally retired.
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
