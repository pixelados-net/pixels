// Package songinfo encodes the TRAX_SONG_INFO outbound response.
package songinfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRAX_SONG_INFO.
const Header uint16 = 3365

// Encode creates an explicit empty song-info list.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(0))
}
