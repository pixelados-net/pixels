// Package permissions contains the USER_PERMISSIONS outbound packet.
package permissions

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the USER_PERMISSIONS packet identifier.
	Header uint16 = 411
)

// Definition describes USER_PERMISSIONS payload fields.
var Definition = codec.Definition{
	codec.Named("clubLevel", codec.Int32Field),
	codec.Named("securityLevel", codec.Int32Field),
	codec.Named("ambassador", codec.BooleanField),
}

// Encode creates a USER_PERMISSIONS packet.
func Encode(clubLevel int32, securityLevel int32, ambassador bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.Int32(clubLevel),
		codec.Int32(securityLevel),
		codec.Bool(ambassador),
	)
}
