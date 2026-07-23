// Package open decodes ACHIEVEMENT_RESOLUTION_OPEN requests.
package open

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ACHIEVEMENT_RESOLUTION_OPEN.
const Header uint16 = 359

// Request stores one resolution furniture request.
type Request struct {
	// ItemID identifies the resolution furniture.
	ItemID int32
	// AchievementID identifies the selected resolution.
	AchievementID int32
}

// Decode returns one resolution furniture request.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	return Request{ItemID: values[0].Int32, AchievementID: values[1].Int32}, nil
}
