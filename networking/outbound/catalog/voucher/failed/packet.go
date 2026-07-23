// Package failed contains the REDEEM_VOUCHER_ERROR outbound packet.
package failed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies REDEEM_VOUCHER_ERROR.
	Header uint16 = 714
	// Invalid identifies an invalid voucher.
	Invalid int32 = iota
	// AlreadyUsed identifies a previously redeemed voucher.
	AlreadyUsed
	// Expired identifies an expired voucher.
	Expired
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field}

// Encode creates a REDEEM_VOUCHER_ERROR packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
