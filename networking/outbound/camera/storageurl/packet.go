// Package storageurl contains CAMERA_STORAGE_URL.
package storageurl

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAMERA_STORAGE_URL.
const Header uint16 = 3696

// Definition describes the durable public URL.
var Definition = codec.Definition{codec.StringField}

// Encode creates a storage URL response.
func Encode(url string) (codec.Packet, error) {
	if url == "" {
		return codec.Packet{}, codec.ErrInvalidField
	}
	return codec.NewPacket(Header, Definition, codec.String(url))
}
