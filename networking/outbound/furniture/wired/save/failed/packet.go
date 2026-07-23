// Package failed encodes the WIRED_ERROR outbound packet.
package failed

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_ERROR packet identifier.
const Header uint16 = 156

// Definition describes WIRED_ERROR.
var Definition = codec.Definition{codec.Named("info", codec.StringField)}

// Encode creates a WIRED validation-error packet.
func Encode(info string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(info))
}
