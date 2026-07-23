// Package expiration contains the GET_CATALOG_PAGE_EXPIRATION inbound packet.
package expiration

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CATALOG_PAGE_EXPIRATION packet identifier.
	Header uint16 = 742
)

// Decode validates a GET_CATALOG_PAGE_EXPIRATION packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
