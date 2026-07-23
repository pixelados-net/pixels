// Package nameapproval encodes CATALOG_APPROVE_NAME_RESULT.
package nameapproval

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CATALOG_APPROVE_NAME_RESULT.
const Header uint16 = 1503

// Encode creates CATALOG_APPROVE_NAME_RESULT.
func Encode(resultCode int32, validationInfo string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(resultCode), codec.String(validationInfo))
}
