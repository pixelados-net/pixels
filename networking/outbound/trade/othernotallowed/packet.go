// Package othernotallowed contains the TRADE_OTHER_NOT_ALLOWED outbound packet.
package othernotallowed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_OTHER_NOT_ALLOWED.
const Header uint16 = 1254

// Encode creates TRADE_OTHER_NOT_ALLOWED.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
