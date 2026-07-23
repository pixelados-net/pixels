// Package keys encodes MYSTERY_BOX_KEYS responses.
package keys

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MYSTERY_BOX_KEYS.
const Header uint16 = 2833

// Definition describes the renderer's string color identifiers.
var Definition = codec.Definition{codec.Named("boxColor", codec.StringField), codec.Named("keyColor", codec.StringField)}

// Encode creates one mystery-box key tracker update.
func Encode(boxColor string, keyColor string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(boxColor), codec.String(keyColor))
}
