// Package apply decodes the WIRED_APPLY_SNAPSHOT inbound packet.
package apply

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_APPLY_SNAPSHOT packet identifier.
const Header uint16 = 3373

// Payload stores one snapshot action request.
type Payload struct {
	// ItemID identifies the match-snapshot action or condition.
	ItemID int32
}

// Definition describes WIRED_APPLY_SNAPSHOT.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode decodes Nitro's confirmed single-item snapshot payload.
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
