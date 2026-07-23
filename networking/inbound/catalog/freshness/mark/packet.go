// Package mark contains the MARK_CATALOG_NEW_ADDITIONS_PAGE_OPENED inbound packet.
package mark

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the MARK_CATALOG_NEW_ADDITIONS_PAGE_OPENED packet identifier.
	Header uint16 = 2150
)

// Decode validates a MARK_CATALOG_NEW_ADDITIONS_PAGE_OPENED packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
