// Package earliest contains the GET_CATALOG_PAGE_WITH_EARLIEST_EXP inbound packet.
package earliest

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CATALOG_PAGE_WITH_EARLIEST_EXP packet identifier.
	Header uint16 = 3135
)

// Decode validates a GET_CATALOG_PAGE_WITH_EARLIEST_EXP packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
