// Package friendcard encodes the friend card shared by Nitro messenger packets.
package friendcard

import "github.com/niflaot/pixels/networking/codec"

// Card contains one Nitro messenger friend projection.
type Card struct {
	// PlayerID identifies the friend.
	PlayerID int64
	// Username stores the visible player name.
	Username string
	// Gender stores Nitro's numeric gender value.
	Gender int32
	// Online reports whether the friend is connected.
	Online bool
	// FollowingAllowed reports whether the friend can be followed now.
	FollowingAllowed bool
	// Look stores the avatar figure string.
	Look string
	// CategoryID stores the friend category, or zero when uncategorized.
	CategoryID int32
	// Motto stores the visible player motto.
	Motto string
	// RealName stores the optional real-name field.
	RealName string
	// LastAccess stores the optional last-access text.
	LastAccess string
	// PersistedMessageUser reports offline-message support.
	PersistedMessageUser bool
	// VIPMember reports VIP membership.
	VIPMember bool
	// PocketHabboUser reports mobile-client presence.
	PocketHabboUser bool
	// Relationship stores the viewer's unilateral relationship marker.
	Relationship uint16
}

// Append appends one friend card in Nitro wire order.
func Append(dst []byte, card Card) ([]byte, error) {
	return codec.AppendPayload(dst, definition,
		codec.Int32(int32(card.PlayerID)), codec.String(card.Username), codec.Int32(card.Gender),
		codec.Bool(card.Online), codec.Bool(card.FollowingAllowed), codec.String(card.Look),
		codec.Int32(card.CategoryID), codec.String(card.Motto), codec.String(card.RealName),
		codec.String(card.LastAccess), codec.Bool(card.PersistedMessageUser), codec.Bool(card.VIPMember),
		codec.Bool(card.PocketHabboUser), codec.Uint16(card.Relationship),
	)
}

// definition describes one friend card in Nitro wire order.
var definition = codec.Definition{
	codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField,
	codec.BooleanField, codec.StringField, codec.Int32Field, codec.StringField,
	codec.StringField, codec.StringField, codec.BooleanField, codec.BooleanField,
	codec.BooleanField, codec.Uint16Field,
}
