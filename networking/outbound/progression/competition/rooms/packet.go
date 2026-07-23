// Package rooms encodes COMPETITION_ROOMS_DATA responses.
package rooms

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_ROOMS_DATA.
const Header uint16 = 3954

// Encode creates one empty or paged competition-room response.
func Encode(goalID int32, pageIndex int32, pageCount int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(goalID), codec.Int32(pageIndex), codec.Int32(pageCount))
}
