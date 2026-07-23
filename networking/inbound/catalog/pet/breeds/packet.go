// Package breeds decodes CATALOG_REQUESET_PET_BREEDS requests.
package breeds

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CATALOG_REQUESET_PET_BREEDS.
const Header uint16 = 1756

// Decode decodes the requested product code.
func Decode(packet codec.Packet) (string, error) {
	if packet.Header != Header {
		return "", codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField})
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
