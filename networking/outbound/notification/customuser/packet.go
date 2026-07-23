// Package customuser encodes CUSTOM_USER_NOTIFICATION furniture notices.
package customuser

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CUSTOM_USER_NOTIFICATION.
const Header uint16 = 909

// Definition describes the native furniture notification code.
var Definition = codec.Definition{codec.Named("code", codec.Int32Field)}

// Encode creates one native furniture notification.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
