// Package concurrentreward decodes GET_CONCURRENT_USERS_REWARD requests.
package concurrentreward

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_CONCURRENT_USERS_REWARD.
const Header uint16 = 3872

// Decode validates one header-only concurrent users reward request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
