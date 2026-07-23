// Package figure contains the USER_FIGURE outbound packet.
package figure

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_FIGURE.
const Header uint16 = 2429

// Definition describes USER_FIGURE fields.
var Definition = codec.Definition{codec.Named("figure", codec.StringField), codec.Named("gender", codec.StringField)}

// Encode creates a USER_FIGURE packet.
func Encode(figure string, gender string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(figure), codec.String(gender))
}
