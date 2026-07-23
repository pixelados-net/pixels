// Package offer contains the retired NEW_USER_EXPERIENCE_GIFT_OFFER packet.
package offer

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NEW_USER_EXPERIENCE_GIFT_OFFER.
const Header uint16 = 3575

// Definition describes the empty retired NUX offer list.
var Definition = codec.Definition{codec.Named("optionGroupCount", codec.Int32Field)}

// Encode creates an empty retired NUX offer list.
//
// Deprecated: the legacy NUX journey is intentionally retired.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(0))
}
