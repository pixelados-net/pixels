// Package voucher contains the catalog voucher redemption event.
package voucher

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one committed voucher redemption.
	Name bus.Name = "catalog.voucher.redeemed"
)

// Payload contains one voucher redemption.
type Payload struct {
	// PlayerID identifies the beneficiary.
	PlayerID int64
	// VoucherID identifies the redeemed voucher.
	VoucherID int64
}
