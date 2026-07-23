// Package completed encodes ACHIEVEMENT_RESOLUTION_COMPLETED responses.
package completed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ACHIEVEMENT_RESOLUTION_COMPLETED.
const Header uint16 = 740

// Encode creates one resolution completion response.
func Encode(itemCode string, badgeCode string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.StringField}, codec.String(itemCode), codec.String(badgeCode))
}
