// Package queue decodes the unused CHANGE_QUEUE request.
package queue

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies CHANGE_QUEUE.
const Header uint16 = 3093

// Definition describes the target queue id.
var Definition = codec.Definition{codec.Named("targetQueue", codec.Int32Field)}

// Decode returns the compatibility-only queue id.
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
