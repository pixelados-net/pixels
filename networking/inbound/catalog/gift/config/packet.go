// Package config contains the GET_GIFT_WRAPPING_CONFIG inbound packet.
package config

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_GIFT_WRAPPING_CONFIG packet identifier.
	Header uint16 = 418
)

// Decode validates a GET_GIFT_WRAPPING_CONFIG packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
