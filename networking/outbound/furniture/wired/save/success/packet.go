// Package success encodes the WIRED_SAVE outbound packet.
package success

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_SAVE packet identifier.
const Header uint16 = 1155

// Encode creates a WIRED save-success packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, nil)
}
