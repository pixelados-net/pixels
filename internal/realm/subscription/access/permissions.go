// Package access contains subscription permission nodes.
package access

import "github.com/niflaot/pixels/internal/permission"

var (
	// VoucherManage permits catalog voucher administration.
	VoucherManage = permission.RegisterNode("catalog.admin.voucher.manage", "")
	// TargetedOfferManage permits targeted-offer administration.
	TargetedOfferManage = permission.RegisterNode("subscription.admin.targeted_offer.manage", "")
	// CalendarManage permits calendar administration.
	CalendarManage = permission.RegisterNode("subscription.admin.calendar.manage", "")
	// ClubOfferManage permits club-offer administration.
	ClubOfferManage = permission.RegisterNode("subscription.admin.club_offer.manage", "")
	// CalendarStaffBypass permits opening unavailable calendar doors.
	CalendarStaffBypass = permission.RegisterNode("subscription.calendar.staff.bypass", "")
	// MembershipGrant permits manual membership mutations.
	MembershipGrant = permission.RegisterNode("subscription.admin.membership.grant", "")
)
