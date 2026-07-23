// Package roominfo contains the ROOM_INFO outbound packet.
package roominfo

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

const (
	// Header is the ROOM_INFO packet identifier.
	Header uint16 = 687
)

// Params contains ROOM_INFO packet data.
type Params struct {
	// RoomEnter reports whether the client should proceed to enter.
	RoomEnter bool
	// Room stores the room data record.
	Room roomcard.Card
	// RoomForward reports whether this response is a forward flow.
	RoomForward bool
	// StaffPick reports whether the room is staff picked.
	StaffPick bool
	// IsGroupMember reports whether the player belongs to the room group.
	IsGroupMember bool
	// AllInRoomMuted reports whether the room is muted.
	AllInRoomMuted bool
	// Moderation stores room moderation settings.
	Moderation ModerationSettings
	// CanMute reports whether the viewer can mute.
	CanMute bool
	// Chat stores room chat settings.
	Chat ChatSettings
}

// ModerationSettings contains room moderation thresholds.
type ModerationSettings struct {
	// AllowMute stores the mute permission level.
	AllowMute int32
	// AllowKick stores the kick permission level.
	AllowKick int32
	// AllowBan stores the ban permission level.
	AllowBan int32
}

// ChatSettings contains room chat settings.
type ChatSettings struct {
	// Mode stores the chat mode.
	Mode int32
	// Weight stores the chat bubble weight.
	Weight int32
	// Speed stores the chat speed.
	Speed int32
	// Distance stores hearing distance.
	Distance int32
	// Protection stores flood protection.
	Protection int32
}

// Definition describes the ROOM_INFO fixed fields.
var Definition = codec.Definition{codec.Named("roomEnter", codec.BooleanField)}

// TailDefinition describes ROOM_INFO fields after room data.
var TailDefinition = codec.Definition{
	codec.Named("roomForward", codec.BooleanField),
	codec.Named("staffPick", codec.BooleanField),
	codec.Named("isGroupMember", codec.BooleanField),
	codec.Named("allInRoomMuted", codec.BooleanField),
}

// ModerationDefinition describes ROOM_INFO moderation fields.
var ModerationDefinition = codec.Definition{
	codec.Named("allowMute", codec.Int32Field),
	codec.Named("allowKick", codec.Int32Field),
	codec.Named("allowBan", codec.Int32Field),
}

// ChatDefinition describes ROOM_INFO chat fields.
var ChatDefinition = codec.Definition{
	codec.Named("canMute", codec.BooleanField),
	codec.Named("mode", codec.Int32Field),
	codec.Named("weight", codec.Int32Field),
	codec.Named("speed", codec.Int32Field),
	codec.Named("distance", codec.Int32Field),
	codec.Named("protection", codec.Int32Field),
}

// Encode creates a ROOM_INFO packet.
func Encode(params Params) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Bool(params.RoomEnter))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = roomcard.Append(payload, params.Room)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendTail(payload, params)
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendTail appends ROOM_INFO fields after room data.
func appendTail(dst []byte, params Params) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, TailDefinition,
		codec.Bool(params.RoomForward),
		codec.Bool(params.StaffPick),
		codec.Bool(params.IsGroupMember),
		codec.Bool(params.AllInRoomMuted),
	)
	if err != nil {
		return dst, err
	}
	dst, err = codec.AppendPayload(dst, ModerationDefinition,
		codec.Int32(params.Moderation.AllowMute),
		codec.Int32(params.Moderation.AllowKick),
		codec.Int32(params.Moderation.AllowBan),
	)
	if err != nil {
		return dst, err
	}

	return codec.AppendPayload(dst, ChatDefinition,
		codec.Bool(params.CanMute),
		codec.Int32(params.Chat.Mode),
		codec.Int32(params.Chat.Weight),
		codec.Int32(params.Chat.Speed),
		codec.Int32(params.Chat.Distance),
		codec.Int32(params.Chat.Protection),
	)
}
