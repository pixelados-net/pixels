// Package purchase contains PURCHASE_PHOTO.
package purchase

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PURCHASE_PHOTO.
const Header uint16 = 2408

// Decode validates the request while ignoring modern optional photo ids.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	return nil
}
