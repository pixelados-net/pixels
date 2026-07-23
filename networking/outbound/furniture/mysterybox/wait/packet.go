// Package wait encodes SHOWMYSTERYBOXWAITMESSAGE responses.
package wait

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SHOWMYSTERYBOXWAITMESSAGE.
const Header uint16 = 3201

// Encode creates a header-only mystery-box wait response.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, codec.Definition{}) }
