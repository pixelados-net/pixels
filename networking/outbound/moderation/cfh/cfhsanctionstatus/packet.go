// Package cfhsanctionstatus contains CFH_SANCTION_STATUS compatibility.
package cfhsanctionstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CFH_SANCTION_STATUS.
const Header uint16 = 2221

// Encode creates the header-only compatibility response.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, nil) }
