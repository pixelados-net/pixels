// Package badgeadd encodes USER_BADGES_ADD responses.
package badgeadd

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BADGES_ADD.
const Header uint16 = 2493

// Encode creates one incremental badge inventory response.
func Encode(badgeID int32, badgeCode string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(badgeID), codec.String(badgeCode))
}
