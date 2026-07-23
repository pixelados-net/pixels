// Package spectator encodes ROOM_SPECTATOR responses.
package spectator

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_SPECTATOR.
const Header uint16 = 1033

// Encode creates a header-only spectator-mode response.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, codec.Definition{}) }
