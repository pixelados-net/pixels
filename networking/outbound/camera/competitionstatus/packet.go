// Package competitionstatus contains COMPETITION_STATUS.
package competitionstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_STATUS.
const Header uint16 = 133

// Definition describes compatibility competition availability.
var Definition = codec.Definition{codec.BooleanField, codec.StringField}

// Encode creates a compatibility competition result.
func Encode(available bool, reason string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(available), codec.String(reason))
}
