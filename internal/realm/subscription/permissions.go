// Package subscription contains subscription realm wiring and permission nodes.
package subscription

import "github.com/niflaot/pixels/internal/realm/subscription/access"

var (
	// VoucherManage permits catalog voucher administration.
	VoucherManage = access.VoucherManage
	// TargetedOfferManage permits targeted-offer administration.
	TargetedOfferManage = access.TargetedOfferManage
	// CalendarManage permits calendar administration.
	CalendarManage = access.CalendarManage
	// ClubOfferManage permits club-offer administration.
	ClubOfferManage = access.ClubOfferManage
	// CalendarStaffBypass permits opening unavailable calendar doors.
	CalendarStaffBypass = access.CalendarStaffBypass
	// MembershipGrant permits manual membership mutations.
	MembershipGrant = access.MembershipGrant
)
