// Package status contains the BUILDERS_CLUB_EXPIRED outbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies BUILDERS_CLUB_EXPIRED.
	Header uint16 = 1452
)

// UNIMPLEMENTED: Builders Club is a discontinued tier without server behavior in Arcturus; neutral values keep Nitro responsive and have no gameplay effect.

// Encode creates a neutral BUILDERS_CLUB_EXPIRED packet.
func Encode() (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.Int32(0), codec.Int32(0), codec.Int32(0), codec.Int32(0))
}
