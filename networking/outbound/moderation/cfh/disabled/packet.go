// Package disabled contains the moderation disabled outbound packet.
package disabled

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation disabled packet.
const Header uint16 = 1651

// Encode creates the header-only moderation disabled packet.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, nil) }
