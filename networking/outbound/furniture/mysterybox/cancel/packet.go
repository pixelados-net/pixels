// Package cancel encodes CANCELMYSTERYBOXWAITMESSAGE responses.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CANCELMYSTERYBOXWAITMESSAGE.
const Header uint16 = 596

// Encode creates a header-only mystery-box cancellation response.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, codec.Definition{}) }
