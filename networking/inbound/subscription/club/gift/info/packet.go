// Package info contains the GET_CLUB_GIFT_INFO inbound packet.
package info

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CLUB_GIFT_INFO packet identifier.
	Header uint16 = 487
)

// Decode validates a GET_CLUB_GIFT_INFO packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
