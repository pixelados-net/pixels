// Package sms contains the DIRECT_SMS_CLUB_BUY outbound packet.
package sms

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies DIRECT_SMS_CLUB_BUY.
	Header uint16 = 195
)

// DEFERRED: Direct SMS billing requires a real carrier provider; this neutral response disables the unavailable purchase path.

// Encode creates an unavailable DIRECT_SMS_CLUB_BUY packet.
func Encode() (codec.Packet, error) {
	definition := codec.Definition{codec.StringField, codec.StringField, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.String(""), codec.String(""), codec.Int32(0))
}
