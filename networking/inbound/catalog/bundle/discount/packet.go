// Package discount contains the GET_BUNDLE_DISCOUNT_RULESET inbound packet.
package discount

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_BUNDLE_DISCOUNT_RULESET packet identifier.
	Header uint16 = 223
)

// Decode validates a GET_BUNDLE_DISCOUNT_RULESET packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
