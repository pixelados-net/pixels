// Package unavailable contains the PURCHASE_NOT_ALLOWED outbound packet.
package unavailable

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PURCHASE_NOT_ALLOWED packet identifier.
	Header uint16 = 3770
	// CodeIllegal identifies an unavailable catalog offer.
	CodeIllegal int32 = 0
	// CodeRequiresClub identifies an offer requiring club membership.
	CodeRequiresClub int32 = 1
)

// Definition describes the PURCHASE_NOT_ALLOWED payload fields.
var Definition = codec.Definition{codec.Named("code", codec.Int32Field)}

// Encode creates a PURCHASE_NOT_ALLOWED packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
