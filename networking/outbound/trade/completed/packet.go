// Package completed contains the TRADE_COMPLETED outbound packet.
package completed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_COMPLETED.
const Header uint16 = 1001

// Encode creates TRADE_COMPLETED.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
