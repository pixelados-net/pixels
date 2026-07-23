// Package ambassador decodes ROOM_AMBASSADOR_ALERT requests.
package ambassador

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_AMBASSADOR_ALERT.
const Header uint16 = 2996

// Definition describes the reported user id carried by Nitro.
var Definition = codec.Definition{codec.Named("userId", codec.Int32Field)}

// Decode returns the reported player id.
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
