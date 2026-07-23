// Package notfound contains the TARGET_OFFER_NOT_FOUND outbound packet.
package notfound

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies TARGET_OFFER_NOT_FOUND.
	Header uint16 = 1237
)

// Encode creates a TARGET_OFFER_NOT_FOUND packet.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
