// Package kickback contains the SCR_GET_KICKBACK_INFO inbound packet.
package kickback

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the SCR_GET_KICKBACK_INFO packet identifier.
	Header uint16 = 869
)

// Decode validates a SCR_GET_KICKBACK_INFO packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
