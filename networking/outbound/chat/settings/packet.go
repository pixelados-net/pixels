// Package settings contains the USER_SETTINGS outbound packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies USER_SETTINGS.
	Header uint16 = 513
)

// Definition describes Nitro user settings including selected chat style.
var Definition = codec.Definition{
	codec.Named("volumeSystem", codec.Int32Field), codec.Named("volumeFurniture", codec.Int32Field),
	codec.Named("volumeTrax", codec.Int32Field), codec.Named("oldChat", codec.BooleanField),
	codec.Named("roomInvitesBlocked", codec.BooleanField), codec.Named("cameraFollowBlocked", codec.BooleanField),
	codec.Named("flags", codec.Int32Field), codec.Named("chatStyle", codec.Int32Field),
}

// Encode creates USER_SETTINGS with neutral defaults and one persisted chat style.
func Encode(chatStyle int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.Int32(100), codec.Int32(100), codec.Int32(100),
		codec.Bool(false), codec.Bool(false), codec.Bool(false), codec.Int32(0), codec.Int32(chatStyle),
	)
}
