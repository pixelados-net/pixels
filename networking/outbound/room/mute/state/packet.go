// Package state contains the ROOM_MUTED outbound packet.
package state

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_MUTED.
	Header uint16 = 2533
)

// Definition describes ROOM_MUTED fields.
var Definition = codec.Definition{codec.Named("muted", codec.BooleanField)}

// Encode creates a ROOM_MUTED packet.
func Encode(muted bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(muted))
}
