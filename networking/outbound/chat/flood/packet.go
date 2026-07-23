// Package flood contains the FLOOD_CONTROL outbound packet.
package flood

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies FLOOD_CONTROL.
	Header uint16 = 566
)

// Definition describes the flood cooldown in seconds.
var Definition = codec.Definition{codec.Named("seconds", codec.Int32Field)}

// Encode creates a FLOOD_CONTROL packet.
func Encode(seconds int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(seconds))
}
