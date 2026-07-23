// Package count contains the BUILDERS_CLUB_QUERY_FURNI_COUNT inbound packet.
package count

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the BUILDERS_CLUB_QUERY_FURNI_COUNT packet identifier.
	Header uint16 = 2529
)

// Decode validates a BUILDERS_CLUB_QUERY_FURNI_COUNT packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
