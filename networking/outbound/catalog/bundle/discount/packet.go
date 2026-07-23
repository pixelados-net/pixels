// Package discount contains the BUNDLE_DISCOUNT_RULESET outbound packet.
package discount

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies BUNDLE_DISCOUNT_RULESET.
	Header uint16 = 2347
)

// Encode creates a BUNDLE_DISCOUNT_RULESET acknowledgement.
func Encode() (codec.Packet, error) {
	return codec.Packet{Header: Header}, nil
}
