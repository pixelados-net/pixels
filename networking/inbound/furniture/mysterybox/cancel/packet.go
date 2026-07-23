// Package cancel decodes MYSTERYBOXWAITINGCANCELEDMESSAGE requests.
package cancel

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MYSTERYBOXWAITINGCANCELEDMESSAGE.
const Header uint16 = 2012

// Definition describes the box owner id.
var Definition = codec.Definition{codec.Named("ownerId", codec.Int32Field)}

// Decode returns the owner id whose pending box wait is canceled.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
