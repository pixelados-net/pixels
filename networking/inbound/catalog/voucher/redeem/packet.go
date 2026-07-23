// Package redeem contains the CATALOG_REDEEM_VOUCHER inbound packet.
package redeem

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CATALOG_REDEEM_VOUCHER.
	Header uint16 = 339
)

// Payload contains voucher redemption fields.
type Payload struct {
	// Code stores the voucher code.
	Code string
}

// Definition describes the payload.
var Definition = codec.Definition{codec.Named("voucherCode", codec.StringField)}

// Decode decodes a voucher redemption.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Code: values[0].String}, nil
}
