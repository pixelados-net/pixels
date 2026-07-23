// Package issuedeleted contains the moderation issuedeleted outbound packet.
package issuedeleted

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation issuedeleted packet.
const Header uint16 = 3192

// Definition describes moderation issuedeleted fields.
var Definition = codec.Definition{
	codec.Named("issueID", codec.StringField),
}

// Encode creates a moderation issuedeleted packet.
func Encode(issueID string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(issueID))
}
