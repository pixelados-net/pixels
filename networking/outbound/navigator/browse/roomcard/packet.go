// Package roomcard encodes navigator room data records.
package roomcard

import "github.com/niflaot/pixels/networking/codec"

const (
	// BitGroup marks room data with group information.
	BitGroup int32 = 2
	// BitAd marks room data with promotion information.
	BitAd int32 = 4
	// BitShowOwner marks room data that should show owner details.
	BitShowOwner int32 = 8
	// BitAllowPets marks room data that allows pets.
	BitAllowPets int32 = 16
	// BitDisplayAd marks room data that should display an entry ad.
	BitDisplayAd int32 = 32
)

// Card contains navigator room data in protocol order.
type Card struct {
	// RoomID identifies the room.
	RoomID int32
	// RoomName stores the room display name.
	RoomName string
	// OwnerID identifies the owner.
	OwnerID int32
	// OwnerName stores the owner display name.
	OwnerName string
	// DoorMode stores the room access mode.
	DoorMode int32
	// UserCount stores current occupancy.
	UserCount int32
	// MaxUserCount stores maximum occupancy.
	MaxUserCount int32
	// Description stores the room description.
	Description string
	// TradeMode stores the room trade setting.
	TradeMode int32
	// Score stores the room score.
	Score int32
	// Ranking stores the navigator ranking.
	Ranking int32
	// CategoryID identifies the navigator category.
	CategoryID int32
	// Tags stores navigator tags.
	Tags []string
	// OfficialRoomPicRef stores an optional room thumbnail reference.
	OfficialRoomPicRef string
	// Group stores optional group data.
	Group *Group
	// Ad stores optional ad data.
	Ad *Ad
	// ShowOwner reports whether the client should show owner data.
	ShowOwner bool
	// AllowPets reports whether pets are allowed.
	AllowPets bool
	// DisplayAd reports whether an entry ad should be displayed.
	DisplayAd bool
}

// Group contains optional navigator room group data.
type Group struct {
	// ID identifies the group.
	ID int32
	// Name stores the group name.
	Name string
	// Badge stores the group badge code.
	Badge string
}

// Ad contains optional navigator room promotion data.
type Ad struct {
	// Name stores the ad title.
	Name string
	// Description stores the ad description.
	Description string
	// ExpiresInMinutes stores remaining ad lifetime.
	ExpiresInMinutes int32
}

// Append appends an encoded room data record to dst.
func Append(dst []byte, card Card) ([]byte, error) {
	dst, err := appendBase(dst, card)
	if err != nil {
		return dst, err
	}

	for _, tag := range card.Tags {
		dst, err = codec.AppendPayload(dst, tagDefinition, codec.String(tag))
		if err != nil {
			return dst, err
		}
	}

	return appendExtras(dst, card)
}

// appendBase appends the fixed room data fields.
func appendBase(dst []byte, card Card) ([]byte, error) {
	return codec.AppendPayload(dst, baseDefinition,
		codec.Int32(card.RoomID),
		codec.String(card.RoomName),
		codec.Int32(card.OwnerID),
		codec.String(card.OwnerName),
		codec.Int32(card.DoorMode),
		codec.Int32(card.UserCount),
		codec.Int32(card.MaxUserCount),
		codec.String(card.Description),
		codec.Int32(card.TradeMode),
		codec.Int32(card.Score),
		codec.Int32(card.Ranking),
		codec.Int32(card.CategoryID),
		codec.Int32(int32(len(card.Tags))),
	)
}

// appendExtras appends bitmask-driven optional fields.
func appendExtras(dst []byte, card Card) ([]byte, error) {
	bitmask := Bitmask(card)
	dst, err := codec.AppendPayload(dst, bitmaskDefinition, codec.Int32(bitmask))
	if err != nil {
		return dst, err
	}
	if card.OfficialRoomPicRef != "" {
		dst, err = codec.AppendPayload(dst, officialDefinition, codec.String(card.OfficialRoomPicRef))
		if err != nil {
			return dst, err
		}
	}
	if card.Group != nil {
		dst, err = appendGroup(dst, *card.Group)
		if err != nil {
			return dst, err
		}
	}
	if card.Ad != nil {
		return appendAd(dst, *card.Ad)
	}

	return dst, nil
}

// appendGroup appends optional group fields.
func appendGroup(dst []byte, group Group) ([]byte, error) {
	return codec.AppendPayload(dst, groupDefinition,
		codec.Int32(group.ID),
		codec.String(group.Name),
		codec.String(group.Badge),
	)
}

// appendAd appends optional promotion fields.
func appendAd(dst []byte, ad Ad) ([]byte, error) {
	return codec.AppendPayload(dst, adDefinition,
		codec.String(ad.Name),
		codec.String(ad.Description),
		codec.Int32(ad.ExpiresInMinutes),
	)
}

// Bitmask returns the protocol bitmask for card optional fields.
func Bitmask(card Card) int32 {
	var bitmask int32
	if card.OfficialRoomPicRef != "" {
		bitmask |= 1
	}
	if card.Group != nil {
		bitmask |= BitGroup
	}
	if card.Ad != nil {
		bitmask |= BitAd
	}
	if card.ShowOwner {
		bitmask |= BitShowOwner
	}
	if card.AllowPets {
		bitmask |= BitAllowPets
	}
	if card.DisplayAd {
		bitmask |= BitDisplayAd
	}

	return bitmask
}

// baseDefinition describes fixed room data fields.
var baseDefinition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("roomName", codec.StringField),
	codec.Named("ownerId", codec.Int32Field),
	codec.Named("ownerName", codec.StringField),
	codec.Named("doorMode", codec.Int32Field),
	codec.Named("userCount", codec.Int32Field),
	codec.Named("maxUserCount", codec.Int32Field),
	codec.Named("description", codec.StringField),
	codec.Named("tradeMode", codec.Int32Field),
	codec.Named("score", codec.Int32Field),
	codec.Named("ranking", codec.Int32Field),
	codec.Named("categoryId", codec.Int32Field),
	codec.Named("tagCount", codec.Int32Field),
}

// tagDefinition describes one room tag field.
var tagDefinition = codec.Definition{codec.Named("tag", codec.StringField)}

// bitmaskDefinition describes the room data bitmask field.
var bitmaskDefinition = codec.Definition{codec.Named("bitMask", codec.Int32Field)}

// officialDefinition describes the optional official room picture field.
var officialDefinition = codec.Definition{codec.Named("officialRoomPicRef", codec.StringField)}

// groupDefinition describes optional group fields.
var groupDefinition = codec.Definition{
	codec.Named("groupId", codec.Int32Field),
	codec.Named("groupName", codec.StringField),
	codec.Named("groupBadge", codec.StringField),
}

// adDefinition describes optional room ad fields.
var adDefinition = codec.Definition{
	codec.Named("adName", codec.StringField),
	codec.Named("adDescription", codec.StringField),
	codec.Named("adExpiresInMin", codec.Int32Field),
}
