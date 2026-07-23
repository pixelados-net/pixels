// Package level encodes TALENT_TRACK_LEVEL responses.
package level

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TALENT_TRACK_LEVEL.
const Header uint16 = 1203

// Encode creates one talent track level summary.
func Encode(name string, current int32, maximum int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field}, codec.String(name), codec.Int32(current), codec.Int32(maximum))
}
