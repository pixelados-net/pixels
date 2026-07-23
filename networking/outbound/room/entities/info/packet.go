// Package info encodes refreshed room-unit profile data.
package info

import "github.com/niflaot/pixels/networking/codec"

// Header is the UNIT_INFO identifier.
const Header uint16 = 3920

// Encode creates a UNIT_INFO packet.
func Encode(unitID int64, figure string, gender string, motto string, achievementScore int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.Int32Field}, codec.Int32(int32(unitID)), codec.String(figure), codec.String(gender), codec.String(motto), codec.Int32(achievementScore))
}
