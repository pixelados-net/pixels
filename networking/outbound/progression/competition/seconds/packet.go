// Package seconds encodes COMPETITION_SECONDS_UNTIL responses.
package seconds

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_SECONDS_UNTIL.
const Header uint16 = 3926

// Encode creates one competition countdown.
func Encode(display string, remaining int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(display), codec.Int32(remaining))
}
