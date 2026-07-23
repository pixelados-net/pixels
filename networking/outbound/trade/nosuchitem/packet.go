// Package nosuchitem contains the TRADE_NO_SUCH_ITEM outbound packet.
package nosuchitem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_NO_SUCH_ITEM.
const Header uint16 = 2873

// Encode creates TRADE_NO_SUCH_ITEM.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
