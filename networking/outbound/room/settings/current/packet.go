// Package current contains the ROOM_SETTINGS outbound packet.
package current

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS.
	Header uint16 = 1498
)

// Params contains the complete editable room settings snapshot.
type Params struct {
	// RoomID identifies the room.
	RoomID int32
	// Name stores the room name.
	Name string
	// Description stores the room description.
	Description string
	// DoorMode stores room access mode.
	DoorMode int32
	// CategoryID identifies the navigator category.
	CategoryID int32
	// MaxUsers stores current capacity.
	MaxUsers int32
	// MaxUsersLimit stores the editor capacity limit.
	MaxUsersLimit int32
	// Tags stores room tags.
	Tags []string
	// TradeMode stores trading mode.
	TradeMode int32
	// AllowPets reports whether pets may enter.
	AllowPets bool
	// AllowPetsEat reports whether pets may eat.
	AllowPetsEat bool
	// AllowWalkthrough reports whether units may overlap.
	AllowWalkthrough bool
	// HideWalls reports whether walls are hidden.
	HideWalls bool
	// WallThickness stores wall thickness.
	WallThickness int32
	// FloorThickness stores floor thickness.
	FloorThickness int32
	// ChatMode stores chat mode.
	ChatMode int32
	// ChatWeight stores bubble weight.
	ChatWeight int32
	// ChatSpeed stores scroll speed.
	ChatSpeed int32
	// ChatDistance stores full hearing range.
	ChatDistance int32
	// ChatProtection stores flood protection.
	ChatProtection int32
	// AllowDynamicCategories reports navigator dynamic-category capability.
	AllowDynamicCategories bool
	// ModerationMute stores mute policy.
	ModerationMute int32
	// ModerationKick stores kick policy.
	ModerationKick int32
	// ModerationBan stores ban policy.
	ModerationBan int32
}

// PrefixDefinition describes fields before variable tags.
var PrefixDefinition = codec.Definition{
	codec.Named("roomId", codec.Int32Field), codec.Named("name", codec.StringField),
	codec.Named("description", codec.StringField), codec.Named("doorMode", codec.Int32Field),
	codec.Named("categoryId", codec.Int32Field), codec.Named("maxUsers", codec.Int32Field),
	codec.Named("maxUsersLimit", codec.Int32Field), codec.Named("tagCount", codec.Int32Field),
}

// SuffixDefinition describes fields after variable tags.
var SuffixDefinition = codec.Definition{
	codec.Named("tradeMode", codec.Int32Field), codec.Named("allowPets", codec.Int32Field),
	codec.Named("allowPetsEat", codec.Int32Field), codec.Named("allowWalkthrough", codec.Int32Field),
	codec.Named("hideWalls", codec.Int32Field), codec.Named("wallThickness", codec.Int32Field),
	codec.Named("floorThickness", codec.Int32Field), codec.Named("chatMode", codec.Int32Field),
	codec.Named("chatWeight", codec.Int32Field), codec.Named("chatSpeed", codec.Int32Field),
	codec.Named("chatDistance", codec.Int32Field), codec.Named("chatProtection", codec.Int32Field),
	codec.Named("allowDynamicCategories", codec.BooleanField), codec.Named("moderationMute", codec.Int32Field),
	codec.Named("moderationKick", codec.Int32Field), codec.Named("moderationBan", codec.Int32Field),
}

// TagDefinition describes one variable tag.
var TagDefinition = codec.Definition{codec.Named("tag", codec.StringField)}

// Encode creates a ROOM_SETTINGS packet.
func Encode(params Params) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, PrefixDefinition, codec.Int32(params.RoomID), codec.String(params.Name), codec.String(params.Description), codec.Int32(params.DoorMode), codec.Int32(params.CategoryID), codec.Int32(params.MaxUsers), codec.Int32(params.MaxUsersLimit), codec.Int32(int32(len(params.Tags))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, tag := range params.Tags {
		payload, err = codec.AppendPayload(payload, TagDefinition, codec.String(tag))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, SuffixDefinition,
		codec.Int32(params.TradeMode), protocolBool(params.AllowPets), protocolBool(params.AllowPetsEat),
		protocolBool(params.AllowWalkthrough), protocolBool(params.HideWalls), codec.Int32(params.WallThickness),
		codec.Int32(params.FloorThickness), codec.Int32(params.ChatMode), codec.Int32(params.ChatWeight),
		codec.Int32(params.ChatSpeed), codec.Int32(params.ChatDistance), codec.Int32(params.ChatProtection),
		codec.Bool(params.AllowDynamicCategories), codec.Int32(params.ModerationMute), codec.Int32(params.ModerationKick),
		codec.Int32(params.ModerationBan))
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// protocolBool converts a boolean to the integer shape required by ROOM_SETTINGS.
func protocolBool(value bool) codec.Value {
	if value {
		return codec.Int32(1)
	}

	return codec.Int32(0)
}
