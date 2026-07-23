// Package halloffame encodes the COMMUNITY_GOAL_HALL_OF_FAME response.
package halloffame

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMMUNITY_GOAL_HALL_OF_FAME.
const Header uint16 = 3005

// Encode creates an empty hall of fame for one goal code.
func Encode(goalCode string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(goalCode), codec.Int32(0))
}
