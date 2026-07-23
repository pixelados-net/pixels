// Package error contains the ROOM_SETTINGS_SAVE_ERROR outbound packet.
package error

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS_SAVE_ERROR.
	Header uint16 = 1555
	// CodeRoomNotFound reports a missing room.
	CodeRoomNotFound int32 = 1
	// CodeNotOwner reports missing settings authorization.
	CodeNotOwner int32 = 2
	// CodeInvalidDoorMode reports unsupported access mode.
	CodeInvalidDoorMode int32 = 3
	// CodeInvalidUserLimit reports unsupported capacity.
	CodeInvalidUserLimit int32 = 4
	// CodeInvalidPassword reports a required password.
	CodeInvalidPassword int32 = 5
	// CodeInvalidCategory reports an unsupported category.
	CodeInvalidCategory int32 = 6
	// CodeInvalidName reports malformed room name.
	CodeInvalidName int32 = 7
	// CodeUnacceptableName reports prohibited room name.
	CodeUnacceptableName int32 = 8
	// CodeInvalidDescription reports malformed description.
	CodeInvalidDescription int32 = 9
	// CodeUnacceptableDescription reports prohibited description.
	CodeUnacceptableDescription int32 = 10
	// CodeInvalidTag reports malformed tags.
	CodeInvalidTag int32 = 11
	// CodeReservedTag reports staff-only tags.
	CodeReservedTag int32 = 12
	// CodeTagTooLong reports excessive tag length.
	CodeTagTooLong int32 = 13
)

// Definition describes ROOM_SETTINGS_SAVE_ERROR fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("code", codec.Int32Field), codec.Named("message", codec.StringField)}

// Encode creates a ROOM_SETTINGS_SAVE_ERROR packet.
func Encode(roomID int32, code int32, message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Int32(code), codec.String(message))
}
