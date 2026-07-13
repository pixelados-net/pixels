// Package notopen contains the TRADE_NOT_OPEN outbound packet.
package notopen

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_NOT_OPEN.
const Header uint16 = 3128

// Encode creates TRADE_NOT_OPEN.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
