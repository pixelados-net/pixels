// Package position encodes GAME_CENTER_IN_ARENA_QUEUE responses.
package position

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_IN_ARENA_QUEUE.
const Header uint16 = 872

// Encode creates one GAME_CENTER_IN_ARENA_QUEUE response.
func Encode(position int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(position))
}
