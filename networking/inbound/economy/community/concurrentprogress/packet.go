// Package concurrentprogress decodes GET_CONCURRENT_USERS_GOAL_PROGRESS requests.
package concurrentprogress

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_CONCURRENT_USERS_GOAL_PROGRESS.
const Header uint16 = 1343

// Decode validates one header-only concurrent users goal request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
