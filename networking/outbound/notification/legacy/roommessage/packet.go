// Package roommessage encodes the retired ROOM_MESSAGE_NOTIFICATION packet.
package roommessage

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_MESSAGE_NOTIFICATION.
const Header uint16 = 1634

// Definition describes the legacy room notice fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("message", codec.StringField), codec.Named("type", codec.Int32Field)}

// Encode creates one compatibility packet.
func Encode(roomID int32, message string, noticeType int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.String(message), codec.Int32(noticeType))
}
