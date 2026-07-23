// Package ok contains the REDEEM_VOUCHER_OK outbound packet.
package ok

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies REDEEM_VOUCHER_OK.
	Header uint16 = 3336
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.StringField}

// Encode creates a REDEEM_VOUCHER_OK packet.
func Encode(productCode string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(productCode))
}
