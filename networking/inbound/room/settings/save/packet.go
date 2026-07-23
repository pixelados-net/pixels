// Package save contains the ROOM_SETTINGS_SAVE inbound packet.
package save

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS_SAVE.
	Header uint16 = 1969
	// MaxTags limits allocation from untrusted tag counts.
	MaxTags int32 = 5
)

// Payload contains unpacked room settings fields.
type Payload struct {
	// RoomID identifies the room.
	RoomID int32
	// Name stores the visible room name.
	Name string
	// Description stores the room description.
	Description string
	// DoorMode stores room access mode.
	DoorMode int32
	// Password stores an optional replacement password.
	Password string
	// MaxUsers stores room capacity.
	MaxUsers int32
	// CategoryID identifies the navigator category.
	CategoryID int32
	// Tags stores the complete room tags.
	Tags []string
	// TradeMode stores room trading mode.
	TradeMode int32
	// AllowPets reports whether pets may enter.
	AllowPets bool
	// AllowPetsEat reports whether pets may eat.
	AllowPetsEat bool
	// AllowWalkthrough reports whether units may overlap.
	AllowWalkthrough bool
	// HideWalls reports whether walls are hidden.
	HideWalls bool
	// WallThickness stores wall render thickness.
	WallThickness int32
	// FloorThickness stores floor render thickness.
	FloorThickness int32
	// ModerationMute stores mute policy.
	ModerationMute int32
	// ModerationKick stores kick policy.
	ModerationKick int32
	// ModerationBan stores ban policy.
	ModerationBan int32
	// ChatMode stores chat display mode.
	ChatMode int32
	// ChatWeight stores bubble weight.
	ChatWeight int32
	// ChatSpeed stores chat scroll speed.
	ChatSpeed int32
	// ChatDistance stores full hearing distance.
	ChatDistance int32
	// ChatProtection stores flood protection.
	ChatProtection int32
}

// PrefixDefinition describes fields before variable tags.
var PrefixDefinition = codec.Definition{
	codec.Named("roomId", codec.Int32Field), codec.Named("name", codec.StringField),
	codec.Named("description", codec.StringField), codec.Named("doorMode", codec.Int32Field),
	codec.Named("password", codec.StringField), codec.Named("maxUsers", codec.Int32Field),
	codec.Named("categoryId", codec.Int32Field), codec.Named("tagCount", codec.Int32Field),
}

// SuffixDefinition describes fields after variable tags.
var SuffixDefinition = codec.Definition{
	codec.Named("tradeMode", codec.Int32Field), codec.Named("allowPets", codec.BooleanField),
	codec.Named("allowPetsEat", codec.BooleanField), codec.Named("allowWalkthrough", codec.BooleanField),
	codec.Named("hideWalls", codec.BooleanField), codec.Named("wallThickness", codec.Int32Field),
	codec.Named("floorThickness", codec.Int32Field), codec.Named("moderationMute", codec.Int32Field),
	codec.Named("moderationKick", codec.Int32Field), codec.Named("moderationBan", codec.Int32Field),
	codec.Named("chatMode", codec.Int32Field), codec.Named("chatWeight", codec.Int32Field),
	codec.Named("chatSpeed", codec.Int32Field), codec.Named("chatDistance", codec.Int32Field),
	codec.Named("chatProtection", codec.Int32Field),
}

// TagDefinition describes one variable tag.
var TagDefinition = codec.Definition{codec.Named("tag", codec.StringField)}

// Decode unpacks a ROOM_SETTINGS_SAVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, PrefixDefinition)
	if err != nil {
		return Payload{}, err
	}
	count := values[7].Int32
	if count < 0 || count > MaxTags {
		return Payload{}, codec.ErrInvalidField
	}
	payload := Payload{RoomID: values[0].Int32, Name: values[1].String, Description: values[2].String, DoorMode: values[3].Int32, Password: values[4].String, MaxUsers: values[5].Int32, CategoryID: values[6].Int32, Tags: make([]string, 0, count)}
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], TagDefinition, rest)
		if err != nil {
			return Payload{}, err
		}
		payload.Tags = append(payload.Tags, values[0].String)
	}
	values, rest, err = codec.DecodePayload(values[:0], SuffixDefinition, rest)
	if err != nil {
		return Payload{}, err
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	applySuffix(&payload, values)

	return payload, nil
}

// applySuffix maps fixed suffix values to a payload.
func applySuffix(payload *Payload, values []codec.Value) {
	payload.TradeMode = values[0].Int32
	payload.AllowPets = values[1].Boolean
	payload.AllowPetsEat = values[2].Boolean
	payload.AllowWalkthrough = values[3].Boolean
	payload.HideWalls = values[4].Boolean
	payload.WallThickness = values[5].Int32
	payload.FloorThickness = values[6].Int32
	payload.ModerationMute = values[7].Int32
	payload.ModerationKick = values[8].Int32
	payload.ModerationBan = values[9].Int32
	payload.ChatMode = values[10].Int32
	payload.ChatWeight = values[11].Int32
	payload.ChatSpeed = values[12].Int32
	payload.ChatDistance = values[13].Int32
	payload.ChatProtection = values[14].Int32
}
