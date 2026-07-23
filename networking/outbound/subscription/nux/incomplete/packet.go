// Package incomplete contains the retired NEW_USER_EXPERIENCE_NOT_COMPLETE packet.
package incomplete

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NEW_USER_EXPERIENCE_NOT_COMPLETE.
const Header uint16 = 3639

// Definition describes the empty retired NUX packet.
var Definition = codec.Definition{}

// Encode creates a retired NEW_USER_EXPERIENCE_NOT_COMPLETE packet.
//
// Deprecated: the legacy NUX journey is intentionally retired.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
