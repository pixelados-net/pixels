// Package basic contains the GET_HABBO_BASIC_MEMBERSHIP_EXTEND_OFFER inbound packet.
package basic

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_HABBO_BASIC_MEMBERSHIP_EXTEND_OFFER packet identifier.
	Header uint16 = 603
)

// Decode validates a GET_HABBO_BASIC_MEMBERSHIP_EXTEND_OFFER packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
