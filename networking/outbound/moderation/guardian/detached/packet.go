// Package detached contains the moderation detached outbound packet.
package detached

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation detached packet.
const Header uint16 = 30

// Encode creates the header-only moderation detached packet.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, nil) }
