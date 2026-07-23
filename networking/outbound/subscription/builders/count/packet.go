// Package count contains the BUILDERS_CLUB_FURNI_COUNT outbound packet.
package count

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies BUILDERS_CLUB_FURNI_COUNT.
	Header uint16 = 3828
)

// Encode creates a neutral BUILDERS_CLUB_FURNI_COUNT packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(0))
}
