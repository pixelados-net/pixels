// Package balance contains the NOT_ENOUGH_BALANCE outbound packet.
package balance

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies NOT_ENOUGH_BALANCE.
	Header uint16 = 3914
)

// Encode creates a NOT_ENOUGH_BALANCE packet.
func Encode(credits bool, points bool, pointsType int32) (codec.Packet, error) {
	definition := codec.Definition{codec.BooleanField, codec.BooleanField, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.Bool(credits), codec.Bool(points), codec.Int32(pointsType))
}
