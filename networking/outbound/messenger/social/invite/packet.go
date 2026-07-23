// Package invite contains the MESSENGER_INVITE outbound packet.
package invite

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_INVITE.
const Header uint16 = 3870

// Encode creates one room invitation.
func Encode(senderID int64, message string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(senderID)), codec.String(message))
}
