// Package info contains GROUP_INFO.
package info

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_INFO.
const Header uint16 = 1702

// Encode creates Nitro's exact non-trailing group information wire shape.
func Encode(group grouprecord.Group, role grouprecord.Role, member bool, pending bool, favorite bool, flag bool, canManage bool) (codec.Packet, error) {
	membershipType := int32(0)
	if member {
		membershipType = 1
	} else if pending {
		membershipType = 2
	}
	isOwner := member && role == grouprecord.Owner
	isAdmin := member && role <= grouprecord.Admin
	pendingCount := int32(0)
	if canManage {
		pendingCount = group.PendingCount
	}
	return codec.NewPacket(Header, codec.Definition{
		codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField,
		codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField,
		codec.BooleanField, codec.BooleanField, codec.StringField, codec.BooleanField, codec.BooleanField, codec.Int32Field,
	}, codec.Int32(int32(group.ID)), codec.Bool(false), codec.Int32(int32(group.State)), codec.String(group.Name), codec.String(group.Description), codec.String(group.BadgeCode),
		codec.Int32(int32(group.HomeRoomID)), codec.String(group.HomeRoomName), codec.Int32(membershipType), codec.Int32(group.MemberCount), codec.Bool(favorite), codec.String(group.CreatedAt.Format("02-01-2006")),
		codec.Bool(isOwner), codec.Bool(isAdmin), codec.String(group.OwnerName), codec.Bool(flag), codec.Bool(group.CanMembersDecorate), codec.Int32(pendingCount))
}
