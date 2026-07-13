// Package confirmation contains the TRADE_CONFIRMATION outbound packet.
package confirmation

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_CONFIRMATION.
const Header uint16 = 2720

// Encode creates TRADE_CONFIRMATION.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
