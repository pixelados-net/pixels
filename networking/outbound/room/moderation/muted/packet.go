// Package muted contains the REMAINING_MUTE outbound packet.
package muted

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies REMAINING_MUTE.
	Header uint16 = 826
)

// Definition describes REMAINING_MUTE fields.
var Definition = codec.Definition{codec.Named("seconds", codec.Int32Field)}

// Encode creates a REMAINING_MUTE packet.
func Encode(seconds int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(seconds))
}
