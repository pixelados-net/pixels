// Package thumbnailupdateresult contains THUMBNAIL_UPDATE_RESULT.
package thumbnailupdateresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies THUMBNAIL_UPDATE_RESULT.
const Header uint16 = 1927

// Definition describes the compatibility result.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Encode creates a legacy thumbnail update result.
func Encode(flatID int32, resultCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(flatID), codec.Int32(resultCode))
}
