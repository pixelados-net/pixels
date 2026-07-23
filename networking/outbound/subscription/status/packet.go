// Package status contains the USER_SUBSCRIPTION outbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies USER_SUBSCRIPTION.
	Header uint16 = 954
)

// State contains client-visible subscription state.
type State struct {
	// ProductName stores the club product code.
	ProductName string
	// DaysToPeriodEnd stores remaining days.
	DaysToPeriodEnd int32
	// MemberPeriods stores completed periods.
	MemberPeriods int32
	// PeriodsAhead stores prepaid future periods.
	PeriodsAhead int32
	// ResponseType identifies the refresh trigger.
	ResponseType int32
	// EverMember reports historical membership.
	EverMember bool
	// VIP reports active VIP status.
	VIP bool
	// PastClubDays stores historical HC days.
	PastClubDays int32
	// PastVIPDays stores historical VIP days.
	PastVIPDays int32
	// MinutesUntilExpiration stores remaining minutes.
	MinutesUntilExpiration int32
}

// Definition describes the required packet fields.
var Definition = codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Encode creates a USER_SUBSCRIPTION packet.
func Encode(state State) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(state.ProductName), codec.Int32(state.DaysToPeriodEnd),
		codec.Int32(state.MemberPeriods), codec.Int32(state.PeriodsAhead), codec.Int32(state.ResponseType),
		codec.Bool(state.EverMember), codec.Bool(state.VIP), codec.Int32(state.PastClubDays),
		codec.Int32(state.PastVIPDays), codec.Int32(state.MinutesUntilExpiration))
}
