// Package younotallowed contains the TRADE_YOU_NOT_ALLOWED outbound packet.
package younotallowed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_YOU_NOT_ALLOWED.
const Header uint16 = 3058

// Encode creates TRADE_YOU_NOT_ALLOWED.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
