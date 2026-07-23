// Package pendingdeleted contains the moderation pendingdeleted outbound packet.
package pendingdeleted

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation pendingdeleted packet.
const Header uint16 = 77

// Encode creates the header-only moderation pendingdeleted packet.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, nil) }
