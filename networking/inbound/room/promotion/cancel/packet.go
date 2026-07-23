// Package cancel decodes CANCEL_ROOM_EVENT requests.
package cancel

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies CANCEL_ROOM_EVENT.
const Header uint16 = 2725

// Definition describes the event id carried by Nitro's composer.
var Definition = codec.Definition{codec.Named("eventId", codec.Int32Field)}

// Decode returns the compatibility-only event id.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return v[0].Int32, nil
}
