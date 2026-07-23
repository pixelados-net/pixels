package model

import "time"

// Voucher contains one redeemable catalog reward.
type Voucher struct {
	// ID identifies the voucher.
	ID int64
	// Code stores the case-insensitive redemption code.
	Code string
	// CostCredits stores credits granted by redemption.
	CostCredits int64
	// CostPoints stores activity points granted by redemption.
	CostPoints int64
	// PointsType identifies the activity-points currency.
	PointsType int32
	// CatalogItemID optionally identifies a granted catalog offer.
	CatalogItemID *int64
	// RedemptionCap optionally limits global redemptions.
	RedemptionCap *int32
	// PerPlayerCap limits redemptions for one player.
	PerPlayerCap int32
	// Enabled reports whether redemption is allowed.
	Enabled bool
	// ExpiresAt stores the optional expiration instant.
	ExpiresAt *time.Time
}

// VoucherRedemption contains one voucher use record.
type VoucherRedemption struct {
	// VoucherID identifies the voucher.
	VoucherID int64
	// PlayerID identifies the beneficiary.
	PlayerID int64
	// RedeemedAt stores the redemption instant.
	RedeemedAt time.Time
}
